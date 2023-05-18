import urllib
import base64

def encode_url(url):
    # URL encoding
    encoded_url = urllib.parse.quote(url)
    # Base64 encoding
    encoded_url = base64.urlsafe_b64encode(encoded_url.encode()).decode()
    return encoded_url
