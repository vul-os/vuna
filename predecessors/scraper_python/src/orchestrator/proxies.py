import requests


def create_proxy_list():
    url = "https://github.com/officialputuid/KangProxy/blob/KangProxy/socks5/socks5.txt"
    response = requests.get(url)
    lines = response.text.split('\n')
    proxy_list = [line.strip() for line in lines if line.strip()]
    return proxy_list

