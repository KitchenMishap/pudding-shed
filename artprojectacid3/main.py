import json

def towerMain():

    print( "Opening source data file..." )
    fi1 = open("Input\\acidblocks3.json")
    jsonFile = json.load(fi1)
    fi1.close()

    print( "First pass: ..." )
    for b, block in enumerate(jsonFile["Blocks"]):
        print( block["Time"] )

towerMain()