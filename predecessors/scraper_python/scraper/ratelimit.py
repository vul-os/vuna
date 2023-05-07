import re
import time
import requests

class WebsiteRateLimiter:
    def __init__(self, url):
        self.url = url
        self.rate_limit_headers = ['X-RateLimit-Limit', 'RateLimit-Limit', 'X-Limit-Request-Limit', 'X-Ratelimit-Limit', 'X-Rate-Limit']
        self.rate_remaining_headers = ['X-RateLimit-Remaining', 'RateLimit-Remaining', 'X-Limit-Request-Remaining', 'X-Ratelimit-Remaining', 'X-Rate-Remaining']
        self.rate_reset_headers = ['X-RateLimit-Reset', 'RateLimit-Reset', 'X-Limit-Request-Reset', 'X-Ratelimit-Reset', 'X-Rate-Reset']
        self.rate_limit = None
        self.rate_remaining = None
        self.rate_reset = None
        self.default_rate_limit = 60
        self.crawl_delay = None

    def _get_rate_limit_info(self):
        response = requests.get(self.url)

        for header, header_list in zip([self.rate_limit, self.rate_remaining, self.rate_reset],
                                       [self.rate_limit_headers, self.rate_remaining_headers, self.rate_reset_headers]):
            for header_name in header_list:
                if header_name in response.headers:
                    header = response.headers[header_name]
                    break
            else:
                header = self.default_rate_limit
            if header is not None:
                header = int(header)
            setattr(self, header_list[0].replace('-', '_').lower(), header)

    def get_crawl_delay(self):
        response_robots = requests.get(self.url + '/robots.txt')
        if response_robots.status_code == 200:
            robots_txt = response_robots.text.lower()
            match = re.search(r'crawl-delay:\s*(\d+)', robots_txt)
            if match:
                self.crawl_delay = int(match.group(1))
                if self.crawl_delay > 0:
                    self.default_rate_limit = 60 / self.crawl_delay

    def __call__(self):
        self.get_crawl_delay()
        self._get_rate_limit_info()
        print('Rate Limit:', self.rate_limit)
        print('Rate Remaining:', self.rate_remaining)
        print('Rate Reset:', self.rate_reset)
        print('Default Rate Limit:', self.default_rate_limit)
        print('Crawl Delay:', self.crawl_delay)

        return self.rate_limit if self.rate_limit is not None else self.default_rate_limit


website_rate_limiter = WebsiteRateLimiter('https://livestainable.co.za')
website_rate_limiter.get_crawl_delay()
rate_limit = website_rate_limiter()
print('Rate Limit:', rate_limit)
