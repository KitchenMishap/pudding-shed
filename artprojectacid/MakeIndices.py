import json

def MakeIndices(inputFile, outputFile):
    fi = open(inputFile)
    ji = json.load(fi)

    jo = {}
    jo["blocks"] = []

    for bi in ji["Blocks"]:
        bo = {}
        epoch = bi["MedianTime"]
        genesis = 1231006505
        jan8_2009 = 1231372800
        if epoch == genesis:
            epoch = jan8_2009   # ToDo MEGABODGE to avoid empty days
        sinceJan8 = epoch - jan8_2009
        day = int(sinceJan8 / 60 / 60 / 24)
        year = int(day / 365)   # This is a bodge approximation
        bo["day"] = day
        bo["year"] = year
        jo["blocks"].append(bo)

    fo = open(outputFile, 'w')
    json.dump(jo,fo,indent=2)

    print("MakeIndices:", len(ji["Blocks"]), "blocks processed")
