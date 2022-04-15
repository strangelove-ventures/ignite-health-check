# Ignite Health Check

Add this container in the same kubernetes pod as a tendermint node to expose a simple listener that will respond with `200 OK` if the node is in sync, otherwise `503 Service Unavailable`

Provide `RPC_ADDRESS` ENV var to change the node address that it will check (default `tcp://localhost:26657`)

Provide `PORT` ENV var to change the port that it will listen on (default `1251`)