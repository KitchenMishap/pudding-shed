import json

def MakeVolumes(inputFile, outputFile):
    fi = open(inputFile)
    ji = json.load(fi)

    jo = {}
    jo["blocks"] = []
    for bi in ji["blocks"]:
        bo = {}

        length = 0.0
        width = 0.0
        thickness = 0.0
        for ci in bi["cuboids"]:
            length += ci["length"]
            width = max(width, ci["width"])
            thickness = max(thickness, ci["thickness"])

        vo = {}
        vo["length"] = length
        vo["width"] = width
        vo["thickness"] = thickness

        bo["volume"] = vo
        jo["blocks"].append(bo)

    fo = open(outputFile, 'w')
    json.dump(jo,fo,indent=2)

    print("MakeVolumes:", len(ji["blocks"]), "blocks processed")
