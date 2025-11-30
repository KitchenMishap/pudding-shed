from spiraltools import *
import json

def towerMain():

    print( "Opening source data file..." )
    fi1 = open("Input\\acidblocks3.json")
    jsonFile = json.load(fi1)
    fi1.close()

    print( "First pass: Import json and create block Instances" )
    for b, block in enumerate(jsonFile["Blocks"]):
        print( block["SizeBytes"] )

towerMain()