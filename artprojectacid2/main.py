from spiraltools import *
import json

fi1 = open("Input\\acidblocks.json")
blocks = json.load(fi1)
fi1.close()

renderer = []

cuboids = []
for i in range(0,10):
    cuboids.append({"length": 10, "width": 20, "thickness": 30})

loop = DayLoop()
for i in range(0,10):
    loop.append(cuboids[i])
loop.process(1.0)
loop.render(renderer, [])

fo = open("Output\\renderspec.json", 'w')
json.dump(renderer,fo,default=vars,indent=2)
