import re


def price_to_float(price):
    if isinstance(price, str):
        if '-' in price:
            price_range = [string_to_float(x) for x in price.split('-')]
            return sum(price_range) / len(price_range)
        else:
            return string_to_float(price)
    if isinstance(price, float):
        return price
    if isinstance(price, int):
        return float(price)
    return None

def max_qty_to_int(max_qty):
    if isinstance(max_qty, str):
        return string_to_int(max_qty)
    if isinstance(max_qty, int):
        return max_qty
    return None

def string_to_float(string):
    # Remove non-digit characters
    cleaned_string = re.sub(r'[^\d.]', '', string)
    # Convert to float
    currency_float = float(cleaned_string)

    return currency_float

def string_to_int(string):
    pattern = r'\b(\d+)\b'
    match = re.search(pattern, string)
    if match:
        number = match.group(1)
        return int(number)
    else:
        return None