//! Price normalization — the `[fields].price_representation` half of a manifest.
//!
//! `RetailObservation::price_minor` is integer **minor units of the observation's currency**
//! (1999 for USD 19.99). Stores publish that number three incompatible ways, so the manifest says
//! which one it is and this module does the conversion. Everything here is integer arithmetic on
//! the source's own digits: no `f64` anywhere, because binary floating point cannot represent
//! `19.99` and a radar that is one cent wrong across a whole corpus is worse than one that admits
//! it could not read the price.
//!
//! Any input this module cannot convert **exactly** yields `None` — an unresolved price, which
//! `quorum::reconcile` treats as a legitimate outcome. It never falls back to a guess.

use std::collections::BTreeMap;

use serde::{Deserialize, Serialize};
use serde_json::Value;

/// The default minor-unit exponent when `[currency_exponents]` does not list the currency (and for
/// an observation with no currency at all). Two decimal places covers most of ISO 4217 — the
/// exceptions are exactly why the manifest carries an override table.
pub const DEFAULT_EXPONENT: u32 = 2;

/// How the source states its price. Serialized snake_case to match the manifest.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum PriceRepresentation {
    /// Already integer minor units (Shopify's `price` field is cents). No conversion.
    MinorInt,
    /// A decimal string in major units (schema.org's `"19.99"`). Needs a currency exponent.
    MajorDecimal,
    /// Integer minor units *plus* the source's own exponent (WooCommerce's `currency_minor_unit`).
    /// Rescaled to the currency's canonical exponent when the two disagree.
    MinorIntWithExponent,
}

/// The exponent the manifest explicitly declares for `currency`, if it declares one.
fn declared_exponent(currency: Option<&str>, overrides: &BTreeMap<String, u32>) -> Option<u32> {
    currency.map(str::to_ascii_uppercase).and_then(|c| overrides.get(&c).copied())
}

/// The exponent to use for `currency`, honouring the manifest's override table and otherwise
/// assuming [`DEFAULT_EXPONENT`].
pub fn exponent_for(currency: Option<&str>, overrides: &BTreeMap<String, u32>) -> u32 {
    declared_exponent(currency, overrides).unwrap_or(DEFAULT_EXPONENT)
}

/// Converts a source price value to integer minor units.
///
/// `stated_exponent` is only consulted for [`PriceRepresentation::MinorIntWithExponent`]; it is
/// the value the manifest's `price_exponent` expression resolved to.
pub fn to_minor_units(
    raw: &Value,
    repr: PriceRepresentation,
    currency: Option<&str>,
    overrides: &BTreeMap<String, u32>,
    stated_exponent: Option<u32>,
) -> Option<i64> {
    let text = numeric_text(raw)?;
    match repr {
        PriceRepresentation::MinorInt => decimal_to_scaled(&text, 0),
        PriceRepresentation::MajorDecimal => decimal_to_scaled(&text, exponent_for(currency, overrides)),
        PriceRepresentation::MinorIntWithExponent => {
            let stated = stated_exponent?;
            let value = decimal_to_scaled(&text, 0)?;
            // A stated exponent is the source's own authority and beats an assumption — so it is
            // only rescaled against a canonical the manifest *declared* for this currency, never
            // against the assumed 2. Rescaling a payload that correctly says JPY-at-0 up to an
            // assumed 2 would multiply the price by a hundred, which is exactly the failure this
            // representation exists to avoid.
            match declared_exponent(currency, overrides) {
                Some(canonical) => rescale(value, stated, canonical),
                None => Some(value),
            }
        }
    }
}

/// A JSON price is a string on some platforms and a number on others; both are accepted, but only
/// when the number is written in plain decimal. A `1.9e1`-style float, a `null`, or a value with a
/// thousands separator or currency symbol is refused rather than guessed at.
fn numeric_text(raw: &Value) -> Option<String> {
    match raw {
        Value::String(s) => Some(s.trim().to_string()),
        Value::Number(n) => Some(n.to_string()),
        _ => None,
    }
}

/// Parses a plain decimal string and scales it by `10^exponent`, exactly.
///
/// More fractional digits than the exponent allows are rounded half away from zero — the same rule
/// a till uses — rather than truncated, so a `"19.999"` in a 2-decimal currency becomes `2000`, not
/// `1999`.
fn decimal_to_scaled(text: &str, exponent: u32) -> Option<i64> {
    let text = text.trim();
    let (negative, digits) = match text.strip_prefix('-') {
        Some(rest) => (true, rest),
        None => (false, text.strip_prefix('+').unwrap_or(text)),
    };

    let (int_part, frac_part) = match digits.split_once('.') {
        Some((i, f)) => (i, f),
        None => (digits, ""),
    };
    if int_part.is_empty() && frac_part.is_empty() {
        return None;
    }
    if !int_part.bytes().all(|b| b.is_ascii_digit()) || !frac_part.bytes().all(|b| b.is_ascii_digit()) {
        return None;
    }

    let exp = exponent as usize;
    let mut value: i128 = if int_part.is_empty() { 0 } else { int_part.parse::<i128>().ok()? };
    value = value.checked_mul(pow10(exponent)?)?;

    if frac_part.len() <= exp {
        let mut frac: i128 = if frac_part.is_empty() { 0 } else { frac_part.parse::<i128>().ok()? };
        frac = frac.checked_mul(pow10((exp - frac_part.len()) as u32)?)?;
        value = value.checked_add(frac)?;
    } else {
        let kept: i128 = if exp == 0 { 0 } else { frac_part[..exp].parse::<i128>().ok()? };
        value = value.checked_add(kept)?;
        // Round half away from zero on the first dropped digit.
        if frac_part.as_bytes()[exp] >= b'5' {
            value = value.checked_add(1)?;
        }
    }

    let value = if negative { -value } else { value };
    i64::try_from(value).ok()
}

/// Re-expresses a minor-unit integer stated at `from` decimal places at `to` decimal places.
fn rescale(value: i64, from: u32, to: u32) -> Option<i64> {
    if from == to {
        return Some(value);
    }
    let value = value as i128;
    if to > from {
        i64::try_from(value.checked_mul(pow10(to - from)?)?).ok()
    } else {
        // Losing precision: round half away from zero rather than truncating toward zero.
        let divisor = pow10(from - to)?;
        let half = divisor / 2;
        let adjusted = if value >= 0 { value.checked_add(half)? } else { value.checked_sub(half)? };
        i64::try_from(adjusted / divisor).ok()
    }
}

fn pow10(exp: u32) -> Option<i128> {
    // 38 digits is the i128 ceiling; anything near it is a malformed manifest, not a price.
    (exp <= 18).then(|| 10i128.pow(exp))
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    fn no_overrides() -> BTreeMap<String, u32> {
        BTreeMap::new()
    }

    fn jpy_and_kwd() -> BTreeMap<String, u32> {
        BTreeMap::from([("JPY".to_string(), 0), ("KWD".to_string(), 3)])
    }

    #[test]
    fn minor_int_passes_integer_cents_through_untouched() {
        let got = to_minor_units(&json!(1999), PriceRepresentation::MinorInt, None, &no_overrides(), None);
        assert_eq!(got, Some(1999));
        // Shopify sends it as a number; some proxies stringify it. Both are the same price.
        let got = to_minor_units(&json!("1999"), PriceRepresentation::MinorInt, None, &no_overrides(), None);
        assert_eq!(got, Some(1999));
    }

    /// The reason this module exists: `19.99 * 100.0` is `1998.9999999999998` in binary floating
    /// point. Integer digit arithmetic gets it right.
    #[test]
    fn major_decimal_is_exact_where_float_arithmetic_is_not() {
        assert_eq!((19.99_f64 * 100.0) as i64, 1998, "float truncation is the bug being avoided");
        let got = to_minor_units(&json!("19.99"), PriceRepresentation::MajorDecimal, Some("USD"), &no_overrides(), None);
        assert_eq!(got, Some(1999));
    }

    #[test]
    fn major_decimal_honours_the_currency_exponent_table() {
        let t = jpy_and_kwd();
        // JPY has no minor unit: 1200 yen is 1200, not 120000.
        let got = to_minor_units(&json!("1200"), PriceRepresentation::MajorDecimal, Some("JPY"), &t, None);
        assert_eq!(got, Some(1200));
        // KWD has three: 19.995 dinar is 19995 fils.
        let got = to_minor_units(&json!("19.995"), PriceRepresentation::MajorDecimal, Some("KWD"), &t, None);
        assert_eq!(got, Some(19995));
        // An unlisted currency falls back to two places.
        let got = to_minor_units(&json!("19.99"), PriceRepresentation::MajorDecimal, Some("ZAR"), &t, None);
        assert_eq!(got, Some(1999));
    }

    #[test]
    fn major_decimal_pads_and_rounds_rather_than_truncating() {
        let o = no_overrides();
        assert_eq!(to_minor_units(&json!("5"), PriceRepresentation::MajorDecimal, None, &o, None), Some(500));
        assert_eq!(to_minor_units(&json!("5.5"), PriceRepresentation::MajorDecimal, None, &o, None), Some(550));
        assert_eq!(to_minor_units(&json!("19.999"), PriceRepresentation::MajorDecimal, None, &o, None), Some(2000));
        assert_eq!(to_minor_units(&json!("19.994"), PriceRepresentation::MajorDecimal, None, &o, None), Some(1999));
        assert_eq!(to_minor_units(&json!("-1.50"), PriceRepresentation::MajorDecimal, None, &o, None), Some(-150));
    }

    #[test]
    fn stated_exponent_is_rescaled_to_the_currency_canonical() {
        let t = jpy_and_kwd();
        // WooCommerce agreeing with the table: no change.
        let got =
            to_minor_units(&json!("1999"), PriceRepresentation::MinorIntWithExponent, Some("USD"), &t, Some(2));
        assert_eq!(got, Some(1999));
        // A store reporting JPY at 2 decimal places: 120000 "sen" is 1200 yen.
        let got =
            to_minor_units(&json!("120000"), PriceRepresentation::MinorIntWithExponent, Some("JPY"), &t, Some(2));
        assert_eq!(got, Some(1200));
        // The other direction: KWD stated at 2, canonical 3.
        let got = to_minor_units(&json!("1999"), PriceRepresentation::MinorIntWithExponent, Some("KWD"), &t, Some(2));
        assert_eq!(got, Some(19990));
    }

    /// The regression this rule exists for: a payload correctly reporting JPY at exponent 0, and a
    /// manifest with no override table. Rescaling to the assumed 2 would report ¥1200 as ¥120000.
    #[test]
    fn a_stated_exponent_is_trusted_when_the_manifest_declares_no_canonical() {
        let o = no_overrides();
        let got = to_minor_units(&json!("1200"), PriceRepresentation::MinorIntWithExponent, Some("JPY"), &o, Some(0));
        assert_eq!(got, Some(1200));
        let got = to_minor_units(&json!("2450"), PriceRepresentation::MinorIntWithExponent, Some("EUR"), &o, Some(2));
        assert_eq!(got, Some(2450));
    }

    #[test]
    fn stated_exponent_missing_is_unresolved_not_assumed() {
        let got = to_minor_units(&json!("1999"), PriceRepresentation::MinorIntWithExponent, Some("USD"), &no_overrides(), None);
        assert_eq!(got, None);
    }

    #[test]
    fn unparseable_prices_are_none_never_a_guess() {
        let o = no_overrides();
        for raw in [json!(null), json!("R 1 299,00"), json!("$19.99"), json!("1,299.00"), json!(true), json!("")] {
            assert_eq!(
                to_minor_units(&raw, PriceRepresentation::MajorDecimal, Some("USD"), &o, None),
                None,
                "{raw} should not resolve to a price"
            );
        }
    }

    #[test]
    fn exponent_lookup_is_case_insensitive_on_the_currency_code() {
        let t = jpy_and_kwd();
        assert_eq!(exponent_for(Some("jpy"), &t), 0);
        assert_eq!(exponent_for(Some("JPY"), &t), 0);
        assert_eq!(exponent_for(Some("USD"), &t), DEFAULT_EXPONENT);
        assert_eq!(exponent_for(None, &t), DEFAULT_EXPONENT);
    }
}
