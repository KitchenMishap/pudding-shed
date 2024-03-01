from spiraltools import *
import json

fi1 = open("Input\\acidblocks.json")
blocks = json.load(fi1)
fi1.close()

renderer = []

yearLoop = YearLoop()

for d in range(0,365):
    dayLoop = DayLoop()
    for b in range(0,144):
        cuboid = Instance(Cube(0,1,1,1,0),[])
        cuboid["length"] = 10
        cuboid["width"] = 20
        cuboid["thickness"] = 30
        dayLoop.append(cuboid)
    dayLoop.process(1.0)
    yearLoop.append(dayLoop)
yearLoop.process(1.0)
yearLoop.render(renderer, [])

fo = open("Output\\renderspec.json", 'w')
json.dump(renderer,fo,default=vars,indent=2)
