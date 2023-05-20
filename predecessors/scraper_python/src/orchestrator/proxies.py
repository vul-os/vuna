import requests


def create_proxy_list(file_path):
    # Send a GET request to retrieve the file contents
    response = requests.get('https://spys.me/proxy.txt')
    # Fetch the proxy list content from the URL
    proxies = response.text.split('\n')

    # Determine the number of header lines
    header_lines = 0
    for line in proxies:
        if line.startswith('IP address:Port'):
            break
        header_lines += 1

    # Remove the header lines
    proxies = proxies[header_lines:]

    proxy_list = []

    # Create proxy URLs for each proxy
    for proxy in proxies:
        proxy_parts = proxy.strip().split(' ')
        ip_port = proxy_parts[0]
        supports_ssl = 'S' in proxy_parts[1]

        # Create the proxy URL
        if not supports_ssl:
            proxy_url = f'http://{ip_port}'
            proxy_list.append(proxy_url)

    return proxy_list