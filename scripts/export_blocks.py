#!/usr/bin/env python3

import argparse

import io
import requests
import os
import sys
import tarfile

parser = argparse.ArgumentParser(description="Exports Nimiq blocks to a tar file")
parser.add_argument("--rpc_url", type=str, required=True, help="RPC server to use")
parser.add_argument(
    "--from", type=int, required=True, dest="from_", help="First block to export"
)
parser.add_argument("--to", type=int, required=True, help="Last block to export")
parser.add_argument("--batch", type=int, default=100, help="Export batch size")
args = parser.parse_args()


class BlockDumper:
    def __init__(self, file):
        self.tar = tarfile.TarFile(mode="w", fileobj=file, format=tarfile.USTAR_FORMAT)

    def __enter__(self):
        return self

    def close(self):
        self.tar.close()

    def __exit__(self, exc_type, exc_value, traceback):
        self.close()

    def add_block(self, number, block_bytes):
        filename = str(number)
        info = tarfile.TarInfo(name=filename)
        info.size = len(block_bytes)
        #info.mtime =
        self.tar.addfile(
            info, fileobj=io.BytesIO(block_bytes)
        )


assert not sys.stdout.isatty()
stdout = os.fdopen(sys.stdout.fileno(), "wb", closefd=False)
dumper = BlockDumper(stdout)

start = args.from_
while True:
    batch = []
    stop = min(start + args.batch, args.to)
    for n in range(start, stop):
        batch.append(
            {
                "jsonrpc": "2.0",
                "id": n,
                "method": "getSerializedBlockByNumber",
                "params": [n],
            }
        )
    if len(batch) == 0:
        break
    print(f"Requesting blocks {start} to {stop}", file=sys.stderr)
    req = requests.post(args.rpc_url, json=batch)
    req.raise_for_status()
    body = req.json()
    start = stop
    sorted(body, key=lambda x: x["id"])
    for block in body:
        result = bytearray.fromhex(block["result"])
        dumper.add_block(block["id"], result)

dumper.close()
stdout.flush()
