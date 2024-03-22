from spiraltools import *
import json

class Block(dict):
    def __init__(self, l, w, t, r, g, b, includeBase, baseR, baseG, baseB):
        self.length = l
        self.minLength = l
        self.width = w
        self.thickness = t
        self.red = r
        self.green = g
        self.blue = b
        self.includeBase = includeBase
        self.baseR = baseR
        self.baseG = baseG
        self.baseB = baseB
        self.baseLength = 0.0
        self.baseWidth = 0.0
        self.introducedTransforms = []

    def render(self, renderer, delegatedTransforms):
        # Firstly, a cube
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

        # Secondly, a slab
        if self.includeBase:
            slabTransform = []
            slabTransform.append(ScaleX(1))     # Much Thinner
            slabTransform.append(ScaleY(math.fabs(self.baseWidth)))
            slabTransform.append(ScaleZ(math.fabs(self.baseLength)))
            colouredSlab = Cube(self.baseR, self.baseG, self.baseB, 1,0)     # Orange
            # Apply an extra transform to base of slab
            slabTransform.append(TranslateX(self.thickness * 0.505))
            # Apply all the introduced transforms
            for introduced in self.introducedTransforms:
                name = introduced.name
                # A block is not distributed over sub-units, so we use the middle
                amount = (introduced.start + introduced.end) / 2
                slabTransform.append(TransformPrimitive(name, amount))
            # Apply all the delegated transforms
            for delegated in delegatedTransforms:
                name = delegated.name
                # A block is not distributed over sub-units, so we use the middle
                amount = (delegated.start + delegated.end) / 2
                slabTransform.append(TransformPrimitive(name, amount))
            positionedSlab = Instance(colouredSlab, slabTransform)
            renderer.append(positionedSlab)

def towerMain():

    daySpacingRatio = 1.0
    yearSpacingRatio = 1.0
    wholeSpacingRatio = 1.0

    print( "Opening source data file...")
    fi1 = open("Input\\acidblocks.json")
    jsonFile = json.load(fi1)
    fi1.close()

    print( "First pass: populate, and measure to percolate up...")
    wholeThing = Loop()
    blk = 1
    blockJson = jsonFile["Blocks"][blk]
    y = 0
    prevY = 0
    d = 5
    prevD = 5
    while y<15:                  # For each year
        yearLoop = Loop()
        while y == prevY:          # For each day in year
            dayLoop = Loop()
            while d == prevD:           # For each block in day
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
                block = Block(length, width, thickness, red, green, blue, False,1.0, 1.0, 1.0)
                dayLoop.append(block)
                blk = blk + 1
                blockJson = jsonFile["Blocks"][blk]
                seconds1970 = blockJson["MedianTime"]
                secondsGenesis = seconds1970 - 1231006505
                daysGenesis = math.floor(secondsGenesis / (24 * 60 * 60))
                yearsGenesis = math.floor(daysGenesis / 365)
                prevY = y
                y = yearsGenesis
                prevD = d
                d = daysGenesis

            dayLoop.complete = True
            prevD = d
            # These measure calls set the following for each loop:
            # minInnerCircumf
            # minLength
            # subUnitsMaxThickness
            # Initially the following are set to these minima for now:
            # innerCircumf
            # length
            # The gist is that sizes of individual blocks will "percolate up" to all the higher level loops
            dayLoop.measure(daySpacingRatio)
            yearLoop.append(dayLoop)

        yearLoop.complete = True
        yearLoop.loopFraction = 1.0
        prevY = y
        yearLoop.measure(yearSpacingRatio)
        wholeThing.append(yearLoop)
    wholeThing.measure(wholeSpacingRatio)
    wholeThing.complete = True
    wholeThing.loopFraction = 1.0

    print("Second pass: Ramped high level measurements percolate down and up...")
    # Enlarge innerCircumf, length based on ramped attributes, measure again
    for y, yearLoop in enumerate(wholeThing.units):
        for d, dayLoop in enumerate(yearLoop.units):
            # Taking account of a "ramped" measurement in yearLoop means it is now potentially bigger than before
            # Here we are "percolating down" this increased measurement to a lower level loops
            dayRadius = yearLoop.maxThicknessRamped(d) / 2.0
            dayInnerCircumf = dayRadius * 2.0 * math.pi
            dayLoop.innerCircumf = max(dayLoop.innerCircumf, dayInnerCircumf)

            # Because the low level measurements have changed as a result, we need to measure again to "percolate"
            # back up to all the higher levels
            dayLoop.measure(daySpacingRatio)
        yearLoop.measure(yearSpacingRatio)
    wholeThing.measure(wholeSpacingRatio)

    print("Third pass: Measure the positions...")
    for y, yearLoop in enumerate(wholeThing.units):
        for d, dayLoop in enumerate(yearLoop.units):
            dayLoop.measurePositions()
        yearLoop.measurePositions()
    wholeThing.measurePositions()

    print("Fourth pass, introduce transforms...")
    for y, yearLoop in enumerate(wholeThing.units):
        for d, dayLoop in enumerate(yearLoop.units):
            dayRadius = dayLoop.innerCircumf / (2.0 * math.pi)
            for b, block in enumerate(dayLoop.units):

                # Transforms introduced at each block
                halfThickness = block.thickness / 2
                block.introducedTransforms.append(SpreadTranslateX(halfThickness, halfThickness))

                # Transforms introduced at each block based on parent's radius for day
                block.introducedTransforms.append(SpreadTranslateX(-dayRadius, -dayRadius))

                # Store some measurements of a "Base" so that block can render a base slab
                block.baseLength = dayRadius * 2.0 * math.pi * block["breadth"]

            # Transforms introduced at each dayLoop based on this dayLoop
            dayLoop.introducedTransforms.append(SpreadRotateY(0,360))
            dayLoop.introducedTransforms.append(SpreadTranslateX(dayRadius, dayRadius))     # Yes another dayRadius

            # Transforms introduced at each dayLoop based on parent's ramped attributes
            yearInnerRadiusRamped = yearLoop.innerCircumfRamped(d) / (2.0 * math.pi)
            dayLoop.introducedTransforms.append(SpreadTranslateX(yearInnerRadiusRamped, yearInnerRadiusRamped))

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

def galaxyMain():

    daySpacingRatio = 1.01
    armSpacingRatio = 1.02
    wholeSpacingRatio = 1.03
    armRatio = 1.0

    print( "Opening source data file...")
    fi1 = open("Input\\acidblocks.json")
    jsonFile = json.load(fi1)
    fi1.close()

    print( "First pass: populate, and measure to percolate up...")
    # No spiralling at this stage
    wholeThing = Loop()
    blk = 1
    blockJson = jsonFile["Blocks"][blk]
    blkCount = len(jsonFile["Blocks"])
    arm = 0
    prevArm = 0
    d = 5
    prevD = 5
    yearsGenesis = 0
    while blk < blkCount - 1:                  # For each arm loop
        armLoop = Loop()
        while arm == prevArm and blk < blkCount - 1:          # For each day in arm loop
            dayLoop = Loop()
            while d == prevD and blk < blkCount - 1:           # For each block in day
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
                if yearsGenesis & 1 == 0:
                    # Bitcoin orange
                    baseR = 255.0 / 256.0
                    baseG = 153.0 / 255.0
                    baseB = 0.0
                else:
                    baseR = 255.0 / 256.0
                    baseG = 255.0 / 256.0
                    baseB = 255.0 / 256.0
                block = Block(length, width, thickness, red, green, blue, baseR, baseG, baseB)
                dayLoop.append(block)
                blk = blk + 1
                blockJson = jsonFile["Blocks"][blk]
                seconds1970 = blockJson["MedianTime"]
                secondsGenesis = seconds1970 - 1231006505
                daysGenesis = math.floor(secondsGenesis / (24 * 60 * 60))
                yearsGenesis = math.floor(daysGenesis / 365.2422)
                prevD = d
                d = daysGenesis

            dayLoop.complete = True
            prevD = d
            # These measure calls set the following for each loop:
            # minInnerCircumf
            # minLength
            # subUnitsMaxThickness
            # Initially the following are set to these minima for now:
            # innerCircumf
            # length
            # The gist is that sizes of individual blocks will "percolate up" to all the higher level loops
            dayLoop.measure(daySpacingRatio)
            armLoop.append(dayLoop)
            # Rather inefficient, but we measure arm every time we add a day
            armLoop.measure(armSpacingRatio)
            if arm == 0:
                targetInnerRadius = armRatio * armLoop.subUnitsMaxThickness
            else:
                prevArmInnerRadius = prevArmLoop.innerCircumf / (2.0 * math.pi)
                targetInnerRadius = prevArmInnerRadius + prevArmLoop.subUnitsMaxThickness
            targetInnerCircumf = 2.0 * math.pi * targetInnerRadius
            if armLoop.minInnerCircumf >= targetInnerCircumf:
                # Arm loop filled, move on to the next one
                armLoop.complete = True
                arm = arm + 1

        prevArm = arm
        prevArmLoop = armLoop
        armLoop.measure(armSpacingRatio)
        wholeThing.append(armLoop)
    wholeThing.measure(wholeSpacingRatio)
    wholeThing.complete = True
    wholeThing.loopFraction = 1.0

    print("Second pass: Ramped high level measurements percolate down and up...")
    # Enlarge innerCircumf, length based on ramped attributes, measure again
    for a, armLoop in enumerate(wholeThing.units):
        for d, dayLoop in enumerate(armLoop.units):
            # Taking account of a "ramped" measurement in armLoop means it is now potentially bigger than before
            # Here we are "percolating down" this increased measurement to a lower level loop
            dayRadius = armLoop.maxThicknessRamped(d) / 2.0 - dayLoop.subUnitsMaxThickness
            dayInnerCircumf = dayRadius * 2.0 * math.pi
            dayLoop.innerCircumf = max(dayLoop.innerCircumf, dayInnerCircumf)

            # Because the low level measurements have changed as a result, we need to measure again to "percolate"
            # back up to all the higher levels
            dayLoop.measure(daySpacingRatio)
        armLoop.measure(armSpacingRatio)
    wholeThing.measure(wholeSpacingRatio)

    print("Third pass: Measure the positions...")
    for a, armLoop in enumerate(wholeThing.units):
        for d, dayLoop in enumerate(armLoop.units):
            dayLoop.measurePositions()
        armLoop.measurePositions()
    wholeThing.measurePositions()

    print("Fourth pass, introduce transforms...")
    print( len(wholeThing.units), " arms...")
    startArmRadius = wholeThing.units[0].startAttr("innerCircumf", False) / (2.0 * math.pi)
    wholeThingRadial = wholeThing.innerCircumf
    ultimateThickness = wholeThing.units[len(wholeThing.units) - 1].endAttr("length", False)
    for a, armLoop in enumerate(wholeThing.units):
        startThickness = armLoop.startAttr("length", False)
        endThickness = armLoop.endAttr("length", False)
        startAxial = (ultimateThickness - startThickness) / 2.0
        endAxial = (ultimateThickness - endThickness) / 2.0
        armRadialStart = armLoop["position"] * wholeThingRadial
        if armLoop.nextUnit is None:
            # Extrapolate
            oneBefore = armLoop.prevUnit
            diff = armLoop["position"] - oneBefore["position"]
            extrapolatedPosition = oneBefore["position"] + (1.0 + armLoop.loopFraction) * diff
            armRadialEnd = extrapolatedPosition * wholeThingRadial
        else:
            armRadialEnd = armLoop.nextUnit["position"] * wholeThingRadial

        startDayDiameter = armLoop.startAttr("length", False)
        endDayDiameter = armLoop.endAttr("length", False)
        startDayInnerRadius = (startDayDiameter - 2.0 * armLoop.startAttr("subUnitsMaxThickness", False)) / 2.0
        endDayInnerRadius = (endDayDiameter - 2.0 * armLoop.endAttr("subUnitsMaxThickness", False)) / 2.0
        for d, dayLoop in enumerate(armLoop.units):
            dayInnerRadius = startDayInnerRadius + dayLoop["position"] * (endDayInnerRadius - startDayInnerRadius)
            armRadiusAtDay = startArmRadius + armRadialStart + dayLoop["position"] * (armRadialEnd - armRadialStart)
            for b, block in enumerate(dayLoop.units):

                # Half block thickness so inside cylinder of dayLoop is smooth
                halfThickness = block.thickness / 2
                block.introducedTransforms.append(SpreadTranslateX(-halfThickness, -halfThickness))

                # Store some measurements of a "Base" so that block can render a base slab
                block.baseLength = dayInnerRadius * 2.0 * math.pi * block["breadth"] * 1.1
                block.baseWidth = armRadiusAtDay * 2.0 * math.pi * dayLoop["breadth"] * 1.8

            # Give dayLoop a radius
            dayLoop.introducedTransforms.append(SpreadTranslateX(dayInnerRadius, dayInnerRadius))

            # Rotation for elements of dayLoop
            dayLoop.introducedTransforms.append(SpreadRotateY(0, 360.0))

        # Expand galaxy based on length of arms
        # So arm radii is driven by wholeThing measurement; we just add the initial radius
        armLoop.introducedTransforms.append(SpreadTranslateX(startArmRadius + armRadialStart, startArmRadius + armRadialEnd))

        # Shift vertically to flatten the top of the spiral
        armLoop.introducedTransforms.append(SpreadTranslateZ(startAxial, endAxial))

        # Rotation for elements of armLoop
        armLoop.introducedTransforms.append(SpreadRotateZ(0, armLoop.loopFraction * 360.0))

    # Transforms introduced at the wholeThing

    print( "Render..." )
    renderer = []
    wholeThing.render(renderer, [])

    print( "Save..." )
    fo = open("Output\\renderspec.json", 'w')
    json.dump(renderer, fo, default=vars, indent=2)

towerMain()
