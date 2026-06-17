#!/usr/bin/env python3
import argparse
import json
import sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.error import HTTPError, URLError
from urllib.request import Request, urlopen


class RelayHandler(BaseHTTPRequestHandler):
    upstream = ""
    timeout = 30

    def do_POST(self):
        length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(length)

        request = Request(
            self.upstream,
            data=body,
            headers={"Content-Type": "application/json"},
            method="POST",
        )

        try:
            with urlopen(request, timeout=self.timeout) as response:
                data = response.read()
                self.send_response(response.status)
                self.send_header("Content-Type", response.headers.get("Content-Type", "application/json"))
                self.send_header("Content-Length", str(len(data)))
                self.end_headers()
                self.wfile.write(data)
        except HTTPError as exc:
            data = exc.read()
            self.send_response(exc.code)
            self.send_header("Content-Type", exc.headers.get("Content-Type", "application/json"))
            self.send_header("Content-Length", str(len(data)))
            self.end_headers()
            self.wfile.write(data)
        except URLError as exc:
            self._write_error(502, f"upstream unavailable: {exc}")
        except Exception as exc:  # pragma: no cover - defensive path
            self._write_error(500, f"relay failure: {exc}")

    def do_GET(self):
        if self.path == "/healthz":
            payload = json.dumps({"ok": True, "upstream": self.upstream}).encode()
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(payload)))
            self.end_headers()
            self.wfile.write(payload)
            return
        self.send_response(405)
        self.end_headers()

    def log_message(self, fmt, *args):
        sys.stdout.write("%s - - [%s] %s\n" % (self.address_string(), self.log_date_time_string(), fmt % args))
        sys.stdout.flush()

    def _write_error(self, status, message):
        payload = json.dumps({"ok": False, "error": message}).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(payload)))
        self.end_headers()
        self.wfile.write(payload)


def main():
    parser = argparse.ArgumentParser(description="Minimal JSON-RPC relay for Sepolia.")
    parser.add_argument("--listen-host", default="127.0.0.1")
    parser.add_argument("--listen-port", type=int, default=18545)
    parser.add_argument("--upstream", required=True)
    parser.add_argument("--timeout", type=int, default=30)
    args = parser.parse_args()

    RelayHandler.upstream = args.upstream
    RelayHandler.timeout = args.timeout

    server = ThreadingHTTPServer((args.listen_host, args.listen_port), RelayHandler)
    print(
        f"relay listening on http://{args.listen_host}:{args.listen_port} -> {args.upstream}",
        flush=True,
    )
    server.serve_forever()


if __name__ == "__main__":
    main()
