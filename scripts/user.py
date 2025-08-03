import sys
import json
import asyncio
import websockets
import requests

API_URL = "http://localhost:15001/api/v1"
WS_URL = "ws://127.0.0.1:8080/ws?user_id={user_id}"

def create_user(username):
    payload = {"username": username}
    resp = requests.post(f"{API_URL}/new_user", json=payload)
    if resp.ok:
        print(f"[+] Created user: {resp.json()}")
    else:
        print(f"[!] Failed to create user: {resp.text}")

async def listen_websocket(user_id):
    url = WS_URL.format(user_id=user_id)
    async with websockets.connect(url, ping_interval=20, ping_timeout=10) as websocket:
        print(f"[+] Connected to WebSocket as user {user_id}")
        try:
            while True:
                message = await websocket.recv()
                print(f"[WS] {message}")
        except websockets.ConnectionClosed as e:
            print(f"[!] WebSocket closed: {e.code} {e.reason}")

def post_tweet(user_id, content):
    payload = {"content": content}
    resp = requests.post(f"{API_URL}/tweet?user={user_id}", json=payload)
    if resp.ok:
        print(f"[+] Tweet posted: {resp.json()}")
    else:
        print(f"[!] Failed to post tweet: {resp.text}")

def main():
    if len(sys.argv) < 4:
        print("Usage: python client.py <username> <create_user:true|false> <websocket:true|false>")
        sys.exit(1)

    username = sys.argv[1]
    create_user_flag = sys.argv[2].lower() == 'true'
    websocket_flag = sys.argv[3].lower() == 'true'

    if create_user_flag:
        create_user(username)

    user_id = input("Enter user ID: ").strip()

    if websocket_flag:
        asyncio.run(listen_websocket(user_id))
    else:
        while True:
            content = input("Enter tweet content: ").strip()
            post_tweet(user_id, content)

if __name__ == "__main__":
    main()
