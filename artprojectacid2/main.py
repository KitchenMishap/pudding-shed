from spiraltools import *
import json


class Block(dict):
    def __init__(self, l, w, t, r, g, b):
        self.length = l
        self.minLength = l
        self.width = w
        self.thickness = t
        self.red = r
        self.green = g
        self.blue = b
        self.introducedTransforms = []

    def render(self, renderer, delegatedTransforms):
        instanceTransform = []
        instanceTransform.append(ScaleX(self.thickness))
        instanceTransform.append(ScaleY(self.width))
        instanceTransform.append(ScaleZ(self.length))
        colouredCube = Cube(self.red, self.green, self.blue, 1,0)
        # Apply all the introduced transforms
        for introduced in self.introducedTransforms:
            name = introduced.name
            # A block is not distributed over sub-units, so we use the middle
            amount = (introduced.start + introduced.end) / 2
            instanceTransform.append(TransformPrimitive(name, amount))
        # Apply all the delegated transforms
        for delegated in delegatedTransforms:
            name = delegated.name
            # A block is not distributed over sub-units, so we use the middle
            amount = (delegated.start + delegated.end) / 2
            instanceTransform.append(TransformPrimitive(name, amount))
        positionedCuboid = Instance(colouredCube, instanceTransform)
        renderer.append(positionedCuboid)

def main():

    daySpacingRatio = 1.0
    yearSpacingRatio = 1.0
    wholeSpacingRatio = 1.0

    print( "Opening source data file...")
    fi1 = open("Input\\acidblocks.json")
    jsonFile = json.load(fi1)
    fi1.close()

    print( "First pass, populate and measure...")
    wholeThing = Loop()
    blk = 0
    for y in range(0,2):
        yearLoop = Loop()
        for d in range(0,365):
            dayLoop = Loop()
            for b in range(0,144):
                blockJson = jsonFile["Blocks"][blk]

                sizeBytes = blockJson["SizeBytes"]
                if sizeBytes >= 16 * 16 * 16:
                    length = math.pow(sizeBytes, 1/3.0)
                    width = math.pow(sizeBytes, 1/3.0)
                    thickness = math.pow(sizeBytes, 1/3.0)
                elif sizeBytes > 16 * 16:
                    width = 16
                    thickness = 16
                    length = sizeBytes / (16 * 16)
                else:
                    length = 1
                    width = 16
                    thickness = sizeBytes / 16

                red = blockJson["ColourByte0"] / 255.0
                green = blockJson["ColourByte1"] / 255.0
                blue = blockJson["ColourByte2"] / 255.0
                block = Block(length, width, thickness, red, green, blue)

                dayLoop.append(block)
                blk = blk + 1

            dayLoop.measure(daySpacingRatio)
            yearLoop.append(dayLoop)

        yearLoop.measure(yearSpacingRatio)
        wholeThing.append(yearLoop)

    wholeThing.measure(wholeSpacingRatio)

    print("Second pass, enlarge innerCircumf, length based on ramped attributes, measure & measure positions")
    # We overwrite things that are based on intended rendering
    for y, yearLoop in enumerate(wholeThing.units):
        for d, dayLoop in enumerate(yearLoop.units):
            dayRadius = yearLoop.maxThicknessRamped(d) / 2.0
            dayInnerCircumf = dayRadius * 2.0 * math.pi
            dayLoop.innerCircumf = max(dayLoop.innerCircumf, dayInnerCircumf)

            dayLoop.measure(daySpacingRatio)
            dayLoop.measurePositions()

        yearLoop.measure(yearSpacingRatio)
        yearLoop.measurePositions()

    wholeThing.measure(wholeSpacingRatio)
    wholeThing.measurePositions()

    print("Third pass, introduce transforms...")
    for y, yearLoop in enumerate(wholeThing.units):
        for d, dayLoop in enumerate(yearLoop.units):
            dayRadius = yearLoop.maxThicknessRamped(d) / 2.0
            for b, block in enumerate(dayLoop.units):

                # Transforms introduced at each block
                halfThickness = block.thickness / 2
                block.introducedTransforms.append(SpreadTranslateX(halfThickness, halfThickness))

                # Transforms introduced at each block based on parent's ramped attributes
                block.introducedTransforms.append(SpreadTranslateX(dayRadius, dayRadius))

            # Transforms introduced at each dayLoop based on this dayLoop
            dayLoop.introducedTransforms.append(SpreadRotateY(0,360))

            # Transforms introduced at each dayLoop based on parent's ramped attributes
            yearMaxThicknessAtDay = yearLoop.maxThicknessRamped(d)
            yearRadiusAtDay = yearMaxThicknessAtDay / 2.0
            dayLoop.introducedTransforms.append(SpreadTranslateX(yearRadiusAtDay, yearRadiusAtDay))

        # Transforms introduced at each yearLoop
        yearLoop.introducedTransforms.append(SpreadRotateZ(0,360))

    # Transforms introduced at the wholeThing
    totalLength = wholeThing.innerCircumf
    wholeThing.introducedTransforms.append(SpreadTranslateZ(0, totalLength))

    print( "Render..." )
    renderer = []
    wholeThing.render(renderer, [])

    print( "Save..." )
    fo = open("Output\\renderspec.json", 'w')
    json.dump(renderer, fo, default=vars, indent=2)


main()
