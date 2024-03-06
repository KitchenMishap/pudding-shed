from spiraltools import *
import json


class Block(dict):
    def __init__(self, l, w, t, r, g, b):
        self.length = l
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
    # First pass: populate and measure()
    fi1 = open("Input\\acidblocks.json")
    jsonFile = json.load(fi1)
    fi1.close()

    renderer = []

    wholeThing = Loop()

    blk = 0
    totalLength = 0.0
    for y in range(0,2):
        yearLoop = Loop()
        for d in range(0,365):
            dayLoop = Loop()
            for b in range(0,144):
                blockJson = jsonFile["Blocks"][blk]
                sizeBytes = blockJson["SizeBytes"]
                size = math.pow(sizeBytes, 1/3.0)
                red = blockJson["ColourByte0"] / 255.0
                green = blockJson["ColourByte1"] / 255.0
                blue = blockJson["ColourByte2"] / 255.0
                block = Block(size, size, size, red, green, blue)

                dayLoop.append(block)
                blk = blk + 1

            dayLoop.measure(1.5)
            yearLoop.append(dayLoop)

        yearLoop.measure(2.0)
        wholeThing.append(yearLoop)

    wholeThing.measure(1.3)

    # Second pass: introduce transforms
    for y, yearLoop in enumerate(wholeThing.units):
        for d, dayLoop in enumerate(yearLoop.units):
            for b, block in enumerate(dayLoop.units):

                # Transforms introduced at each block
                halfThickness = block.thickness / 2
                block.introducedTransforms.append(SpreadTranslateX(halfThickness, halfThickness))

            # Transforms introduced at each dayLoop
            dayInnerRadius = dayLoop.innerRadius()
            dayLoop.introducedTransforms.append(SpreadTranslateX(dayInnerRadius, dayInnerRadius))
            dayLoop.introducedTransforms.append(SpreadRotateY(0,360))

        # Transforms introduced at each yearLoop
        yearInnerRadius = yearLoop.innerRadius()
        yearMaxThickness = yearLoop.maxThickness
        yearRadius = yearInnerRadius + yearMaxThickness / 2
        yearLoop.introducedTransforms.append(SpreadTranslateX(yearRadius, yearRadius))
        yearLoop.introducedTransforms.append(SpreadRotateZ(0,360))
        totalLength += yearLoop.length()

    # Transforms introduced at the wholeThing
    wholeThing.introducedTransforms.append(SpreadTranslateZ(0, totalLength * 1.3))

    wholeThing.render(renderer, [])

    fo = open("Output\\renderspec.json", 'w')
    json.dump(renderer, fo, default=vars, indent=2)


main()
