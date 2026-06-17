const http = require("node:http");
const https = require("node:https");
const { URL } = require("node:url");

const upstream = process.env.UPSTREAM_RPC_URL;
const listenHost = process.env.RELAY_LISTEN_HOST || "127.0.0.1";
const listenPort = Number(process.env.RELAY_LISTEN_PORT || "18545");
const requestTimeoutMs = Number(process.env.RELAY_TIMEOUT_MS || "30000");

if (!upstream) {
  console.error("UPSTREAM_RPC_URL is required");
  process.exit(1);
}

const upstreamUrl = new URL(upstream);

const server = http.createServer((req, res) => {
  if (req.method === "GET" && req.url === "/healthz") {
    const body = JSON.stringify({ ok: true, upstream });
    res.writeHead(200, {
      "Content-Type": "application/json",
      "Content-Length": Buffer.byteLength(body),
      Connection: "close",
    });
    res.end(body);
    return;
  }

  if (req.method !== "POST") {
    res.writeHead(405, { Connection: "close" });
    res.end();
    return;
  }

  const chunks = [];
  req.on("data", (chunk) => chunks.push(chunk));
  req.on("end", () => {
    const body = Buffer.concat(chunks);
    const proxyReq = https.request(
      {
        protocol: upstreamUrl.protocol,
        hostname: upstreamUrl.hostname,
        port: upstreamUrl.port || 443,
        path: `${upstreamUrl.pathname}${upstreamUrl.search}`,
        method: "POST",
        headers: {
          "Content-Type": req.headers["content-type"] || "application/json",
          "Content-Length": body.length,
          Connection: "close",
        },
        timeout: requestTimeoutMs,
      },
      (proxyRes) => {
        res.writeHead(proxyRes.statusCode || 502, {
          "Content-Type": proxyRes.headers["content-type"] || "application/json",
          Connection: "close",
        });
        proxyRes.pipe(res);
      }
    );

    proxyReq.on("timeout", () => {
      proxyReq.destroy(new Error("upstream timeout"));
    });

    proxyReq.on("error", (error) => {
      const payload = JSON.stringify({ ok: false, error: error.message });
      res.writeHead(502, {
        "Content-Type": "application/json",
        "Content-Length": Buffer.byteLength(payload),
        Connection: "close",
      });
      res.end(payload);
    });

    proxyReq.end(body);
  });

  req.on("error", (error) => {
    const payload = JSON.stringify({ ok: false, error: error.message });
    res.writeHead(400, {
      "Content-Type": "application/json",
      "Content-Length": Buffer.byteLength(payload),
      Connection: "close",
    });
    res.end(payload);
  });
});

server.listen(listenPort, listenHost, () => {
  console.log(`relay listening on http://${listenHost}:${listenPort} -> ${upstream}`);
});
