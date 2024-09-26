from discord.ext import commands, tasks
import discord
import datetime





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
async def set_reminder(ctx, user: discord.Member):
    my_task.start(user.id)



@tasks.loop(seconds=10)
async def my_task(user_id):
    user = bot.get_user(user_id)
    message = "Máš příští schůzi. Připrav si jí v čas. \n\nreaguj na tuto zprávu aby jsi vypnul připomínky."
    msg = await user.send(message)
    await msg.add_reaction('💀')

bot.run(token)
