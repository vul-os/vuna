//! Body-size cap. Kept as a pure function over any [`Read`] so it is unit-testable with an
//! in-memory `Cursor` — no network, no `reqwest` types, no async runtime touch this module.

use std::io::{self, Read};

/// Reads up to `max_bytes` from `reader`. If the stream still has data beyond that point the body
/// is considered oversized: rather than silently truncating (which would hand an extractor a
/// corrupt document it can't tell apart from a genuinely short page) this returns an error so the
/// caller can surface a clean [`vuna_core::Error::Fetch`].
pub fn read_capped<R: Read>(reader: R, max_bytes: usize) -> io::Result<Vec<u8>> {
    let mut buf = Vec::with_capacity(max_bytes.min(64 * 1024));
    // Read one byte past the cap so we can tell "exactly at the cap" from "oversized" without
    // needing the underlying reader to expose a length up front (chunked transfer encoding etc.
    // often doesn't have one).
    let mut limited = reader.take(max_bytes as u64 + 1);
    limited.read_to_end(&mut buf)?;
    if buf.len() > max_bytes {
        return Err(io::Error::other(format!("body exceeds cap of {max_bytes} bytes")));
    }
    Ok(buf)
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::io::Cursor;

    #[test]
    fn body_within_cap_is_returned_whole() {
        let data = vec![1u8; 100];
        let out = read_capped(Cursor::new(data.clone()), 1024).unwrap();
        assert_eq!(out, data);
    }

    #[test]
    fn body_exactly_at_cap_is_ok() {
        let out = read_capped(Cursor::new(vec![7u8; 10]), 10).unwrap();
        assert_eq!(out.len(), 10);
    }

    #[test]
    fn body_one_byte_over_cap_is_rejected() {
        let err = read_capped(Cursor::new(vec![7u8; 11]), 10).unwrap_err();
        assert_eq!(err.kind(), io::ErrorKind::Other);
    }

    #[test]
    fn body_far_over_cap_is_rejected() {
        let err = read_capped(Cursor::new(vec![7u8; 10_000]), 10).unwrap_err();
        assert_eq!(err.kind(), io::ErrorKind::Other);
    }

    #[test]
    fn empty_body_is_ok() {
        let out = read_capped(Cursor::new(Vec::<u8>::new()), 10).unwrap();
        assert!(out.is_empty());
    }

    #[test]
    fn zero_cap_rejects_any_nonempty_body() {
        let err = read_capped(Cursor::new(vec![1u8]), 0).unwrap_err();
        assert_eq!(err.kind(), io::ErrorKind::Other);
        assert!(read_capped(Cursor::new(Vec::<u8>::new()), 0).is_ok());
    }
}
