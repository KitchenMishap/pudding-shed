import json
import colorsys

def MakeCuboids(inputFile, outputFile):
    fi = open(inputFile)
    ji = json.load(fi)

    jo = {}
    jo["blocks"] = []
    for bi in ji["Blocks"]:
        bo = {}
        bo["cuboids"] = []

        cu1 = {}
        cu2 = {}
        cu3 = {}
        cu4 = {}

        # Colour
        hue = bi["ColourByte0"] / 255
        sat = 1.0
        val = 1.0
        (r,g,b) = colorsys.hsv_to_rgb(hue,sat,val)

        cu1["r"] = r
        cu1["g"] = g
        cu1["b"] = b
        cu1["a"] = 1

        # Size / shape
        sizeBytes = bi["SizeBytes"]

        # Volume is always sizeBytes
        if sizeBytes < 16*16:
            # Below 16x16 is indicated as a rectangular slab
            cu1["length"] = 1
            cu1["width"] = 16
            cu1["thickness"] = sizeBytes / 16
        elif sizeBytes < 16*16*16:
            # 16x16..16x16x16 is indicated as a slab with 16x16 cross section
            cu1["length"] = sizeBytes / (16 * 16)
            cu1["width"] = 16
            cu1["thickness"] = 16
        else:
            # Above 16x16x16 is indicated as a cube
            side = pow(sizeBytes, 1.0/3.0)
            cu1["length"] = side
            cu1["width"] = side
            cu1["thickness"] = side

        # Other stuff
        cu1["hism"] = 0
        bo["cuboids"].append(cu1)

        jo["blocks"].append(bo)

    fo = open(outputFile, 'w')
    json.dump(jo,fo,indent=2)

    print("MakeCuboids:", len(ji["Blocks"]), "blocks processed")
