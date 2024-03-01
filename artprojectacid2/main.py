from spiraltools import *
import json

fi1 = open("Input\\acidblocks.json")
blocks = json.load(fi1)
fi1.close()

renderer = []

loop = DayLoop()
for i in range(0,10):
    cuboid = Instance(Cube(0,1,1,1,0),[])
    cuboid["length"] = 10
    cuboid["width"] = 20
    cuboid["thickness"] = 30
    loop.append(cuboid)

loop.process(1.0)
loop.render(renderer, [])

fo = open("Output\\renderspec.json", 'w')
json.dump(renderer,fo,default=vars,indent=2)
