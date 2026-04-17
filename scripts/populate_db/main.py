import argparse

import requests
import urllib3
from populate_test import populate_test

urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

BASE_URL = "https://localhost/api"
PROXY_URL = "http://127.0.0.1:8080"


if __name__ == "__main__":
    parser = argparse.ArgumentParser()

    parser.add_argument(
        "--base-url",
        type=str,
        default=BASE_URL,
        help=f"Base URL for the Kryvea API (default: {BASE_URL})",
    )
    parser.add_argument(
        "--enable-proxy",
        type=bool,
        default=False,
        help="Enable proxy",
    )
    parser.add_argument(
        "--proxy",
        type=str,
        default=PROXY_URL,
        help=f"Proxy URL (default: {PROXY_URL})",
    )
    parser.add_argument(
        "--username",
        type=str,
        default="kryvea",
        help="Admin username (default: kryvea)",
    )
    parser.add_argument(
        "--password",
        type=str,
        default="Kryvea123!",
        help="Admin password (default: Kryvea123!)",
    )
    args = parser.parse_args()

    session = requests.Session()
    session.verify = False
    session.proxies = {
        "http": args.proxy,
        "https": args.proxy,
    }
    if not args.enable_proxy:
        session.proxies = {}

    # populate db with test data
    try:
        data = populate_test(session, args.base_url, args.username, args.password)
    except Exception as e:
        print(f"Error populating database: {e}")
