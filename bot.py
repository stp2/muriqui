from discord.ext import commands, tasks
import discord
from datetime import datetime
import time


with open('token.txt', 'r') as f:
    token = f.read()

bot = commands.Bot(command_prefix='!', intents=discord.Intents.all())


@bot.command()
async def hello(ctx):
    await ctx.send('hella')

@bot.event
async def on_reaction_add(reaction, user):
    if user == bot.user:
        return
    else:
        my_task.cancel()


@bot.command()
async def set_reminder(ctx, user: discord.Member, wait_time):
        dt_now = datetime.now()

        dt_list = wait_time.split('.')
        dt_new = datetime(int(dt_list[2]), int(dt_list[1]), int(dt_list[0]), 18)
        duration = dt_new - dt_now
        print(round(duration.total_seconds()))
        if round(duration.total_seconds()) <= 0:
            await ctx.send("```You can't pick that time!```")
        '''
        time.sleep(round(duration.total_seconds()))
        my_task.start(user.id)
        '''



@tasks.loop(hours=24)
async def my_task(user_id):
    user = bot.get_user(user_id)
    message = "```Máš příští schůzi. Připrav si jí v čas. \n\nreaguj na tuto zprávu aby jsi vypnul připomínky.```"
    msg = await user.send(message)
    await msg.add_reaction('💀')



bot.run(token)
