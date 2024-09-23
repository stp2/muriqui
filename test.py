#!/usr/bin/env python3
# This example requires the 'message_content' intent.

import discord
import json
import asyncio

token : str

with open('config.json') as f:
    config = json.load(f)
    token = config['token']

intents = discord.Intents.default()
intents.members = True

client = discord.Client(intents=intents)

@client.event
async def on_ready():
    print(f'We have logged in as {client.user}')

async def main():
    task = asyncio.create_task(client.login(token))
    await task
    print(client.users)
    user = client.get_user(0) # Replace 0 with the user ID
    print(user)
    await user.send('Hello!')
    await client.connect()

asyncio.run(main())