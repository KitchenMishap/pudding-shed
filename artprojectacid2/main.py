from spiraltools import *
import json

fi1 = open("Input\\acidblocks.json")
blocks = json.load(fi1)
fi1.close()

renderer = []

cuboids = []
for i in range(0,10):
    cuboids.append({"length": 10, "width": 20, "thickness": 30})

loop = Loop()
for i in range(0,10):
    loop.append(cuboids[i])

loop.process(1.0)
innerRadius = loop.innerRadius()
for i in range(0,10):
    position = loop.units[i]["position"]
    transforms = []
    transforms.append(ScaleX(loop.units[i]["thickness"]))
    transforms.append(ScaleY(loop.units[i]["width"]))
    transforms.append(ScaleZ(loop.units[i]["length"]))
    radius = innerRadius + loop.units[i]["thickness"] / 2
    transforms.append(TranslateX(radius))
    angle = position * 360
    transforms.append(RotateY(angle))
    renderer.append( Instance( Cube(0,1,0,1,0), transforms ) )

fo = open("Output\\renderspec.json", 'w')
json.dump(renderer,fo,default=vars,indent=2)
