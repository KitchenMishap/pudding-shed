from spiraltools import *
import json

fi1 = open("Input\\acidblocks.json")
blocks = json.load(fi1)
fi1.close()

renderer = []

renderer.append( Instance( Cube(1,0,0,1,0), [RotateX(45)] ) )

fo = open("Output\\renderspec.json", 'w')
json.dump(renderer,fo,default=vars,indent=2)
