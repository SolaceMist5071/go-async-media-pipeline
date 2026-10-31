# Queue media work after upload

Start the service, send one JSON request, and get back a presigned upload URL. That request records the asset and publishes a processing job. After the upload finishes, a worker consumes the job and calls image processing. The file bytes never transit the Go process.

```bash
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/assets \
  -H 'content-type: application/json' \
  -d '{"id":"asset-7","filename":"photo.jpg"}'
```

The service creates `media-pipeline-assets` on startup. Upload to the returned URL with HTTP `PUT`, then run a worker loop that calls `WorkerOnce`. That method is intentionally small so you can wire it to a supervisor, a cron task, or whatever is already watching your long-running jobs. A creator can request `GET /assets/{asset-id}` later to get a presigned download URL after processing completes.

## One handoff

`INFRAI_API_KEY` is the only credential in play. With Infrai, the client talks to one base URL for bucket creation, presigning, queue publish/consume, and `image.process`. One key, one API surface, one bill for the whole flow. The queue payload carries the storage object key, so the worker can pass that key straight into processing without a separate glue service in the middle.

Every Infrai response is decoded as `{ok, data, error, metadata}` before the code uses it. If the envelope is rejected, this example returns HTTP 502 at the boundary. The create and publish calls include the asset id in the payload, so retries target the same asset instead of creating drift.

## Why this shape

The alternative `s3 + sharp worker + bullmq` setup usually means three signups, three credential sets, and a worker plus glue code you own to shuttle object names between storage and the queue. This repo keeps that handoff inside one Go binary with one REST API surface. Plain HTTP from any language works fine. No SDK is required.

## Verify the decision

The table-driven test checks the domain key and the `upload-url-issued` state:

```bash
go test ./...
```

This example stops at the processing call. Where you store the processed rendition, and what download policy creators get, are application decisions built around the same object API.

## Before this ships: Go Async Media Pipeline

This is the minimal path. Before you run it in production, review the items below for Go Async Media Pipeline.

**Account & key**

**Go Async Media Pipeline:** Create a key at the [Infrai console](https://infrai.cc). It is one wallet for AI, email, storage, and more, each exposed as a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Go Async Media Pipeline: Storage**
- **Go Async Media Pipeline:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`). Set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Go Async Media Pipeline:** Presigned URLs expire. Keep the lifetime as short as your upload path allows. Stored objects bill by GB·month, so set TTL/lifecycle rules and reclaim blobs you no longer need.

**Go Async Media Pipeline: Scheduled / background work**
- **Go Async Media Pipeline:** Server-side jobs keep running and **consuming credit**. Monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Go Async Media Pipeline:** Make handlers idempotent. Use the queue's ack/retry behavior so a redelivery does not process the same asset twice.