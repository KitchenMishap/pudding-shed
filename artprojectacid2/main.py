from spiraltools import *
import json

class Block(dict):
    def __init__(self, glwBefore, glwAfter, l, w, t, r, g, b, includeBase, baseR, baseG, baseB):
        self.gapLengthWeightBefore = glwBefore  # Spare space is distributed between blocks according to glw,
        self.gapLengthWeightAfter = glwAfter    # giving PARTIAL indication of of delta-time between blocks
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
        instanceTransform.append(ScaleZ(self.minLength))
        colouredCube = Cube(self.red, self.green, self.blue, 1,0)
        # Apply all the introduced transforms
        for introduced in self.introducedTransforms:
            name = introduced.name
            amount = introduced.start
            instanceTransform.append(TransformPrimitive(name, amount))
        # Apply all the delegated transforms
        for delegated in delegatedTransforms:
            name = delegated.name
            amount = delegated.start
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
                amount = introduced.start
                slabTransform.append(TransformPrimitive(name, amount))
            # Apply all the delegated transforms
            for delegated in delegatedTransforms:
                name = delegated.name
                amount = delegated.start
                slabTransform.append(TransformPrimitive(name, amount))
            positionedSlab = Instance(colouredSlab, slabTransform)
            renderer.append(positionedSlab)

def markBlock(json, block):
    json["Blocks"][block-2]["ColourByte0"] = 255    # red before
    json["Blocks"][block-2]["ColourByte1"] = 0
    json["Blocks"][block-2]["ColourByte2"] = 0
    json["Blocks"][block-1]["ColourByte0"] = 255    # red before
    json["Blocks"][block-1]["ColourByte1"] = 0
    json["Blocks"][block-1]["ColourByte2"] = 0
    json["Blocks"][block]["ColourByte0"] = 0    # blue block
    json["Blocks"][block]["ColourByte1"] = 0
    json["Blocks"][block]["ColourByte2"] = 255
    json["Blocks"][block+1]["ColourByte0"] = 255    # red after
    json["Blocks"][block+1]["ColourByte1"] = 0
    json["Blocks"][block+1]["ColourByte2"] = 0
    json["Blocks"][block+2]["ColourByte0"] = 255    # red after
    json["Blocks"][block+2]["ColourByte1"] = 0
    json["Blocks"][block+2]["ColourByte2"] = 0

TIMEZERO = 1230768000   # Midnight Jan 1 2009
def secondsIntoDay(timestamp):
    return (timestamp - TIMEZERO) % (24 * 60 * 60)
def dayNumber(timestamp):
    return math.floor(float(timestamp - TIMEZERO) / (24 * 60 * 60))
def yearNumber(timestamp):
    return math.floor(dayNumber(timestamp) / 365)

def towerMain():

    daySpacingRatio = 1.1
    yearSpacingRatio = 1.3
    centurySpacingRatio = 1.0

    print( "Opening source data file...")
    fi1 = open("Input\\acidblocks.json")
    jsonFile = json.load(fi1)
    fi1.close()

    print( "Making difficulty adjustment blocks black... and grey for the next 144 blocks")
    for b, block in enumerate(jsonFile["Blocks"]):
        sinceDifficulty = b % 2016
        if b >= 2016 and sinceDifficulty < 144:
            jsonFile["Blocks"][b]["ColourByte0"] = 128
            jsonFile["Blocks"][b]["ColourByte1"] = 128
            jsonFile["Blocks"][b]["ColourByte2"] = 128
            if sinceDifficulty == 0:
                jsonFile["Blocks"][b]["ColourByte0"] = 0
                jsonFile["Blocks"][b]["ColourByte1"] = 0
                jsonFile["Blocks"][b]["ColourByte2"] = 0

    print( "Marking significant blocks as blue... (and two red blocks either side)")
    markBlock(jsonFile, 170)      # First transaction, 10 btc satoshi to finney
    markBlock(jsonFile, 57043)    # Pizza transaction

    print( "First pass: populate, and measure to percolate up...")
    centuryLoop = Loop()
    blk = 1     # Don't start with block 0 as that is a block that's alone in a day and confuses things
    blockJson = jsonFile["Blocks"][blk]
    timestamp = blockJson["MedianTime"]
    secondOfDay = secondsIntoDay(timestamp)
    day = dayNumber(timestamp)
    year = yearNumber(timestamp)
    firstBlockOfDay = True
    prevYear = year
    prevDay = day
    prevSecond = secondOfDay - 10 * 60
    prevBlock = None
    end = False
    while not end:                  # For each year
        yearLoop = Loop()
        while year == prevYear and not end:          # For each day in year
            dayLoop = Loop()
            while (day == prevDay or firstBlockOfDay) and not end:           # For each block in day
                # PART 1) Do the stuff for the current block
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
                dayLoop.endAngle = 360.0 * (float(secondOfDay) / (24.0 * 60.0 * 60.0))  # Overwritten unless last block
                if firstBlockOfDay:
                    dayLoop.startAngle = dayLoop.endAngle
                gapLengthWeight = max(secondOfDay - prevSecond, 60)     # Minimum of a minute, as early blocks have medianTime spacing zero?
                block = Block(gapLengthWeight, 0.0, length, width, thickness, red, green, blue, False, 1.0, 1.0, 1.0)
                dayLoop.append(block)
                if not (prevBlock is None):
                    prevBlock["gapLengthWeightAfter"] = gapLengthWeight
                # PART 2) Move onto the next block
                prevBlock = block
                prevSecond = secondOfDay
                prevDay = day
                prevYear = year
                blk = blk + 1
                if blk < len(jsonFile["Blocks"]):
                    blockJson = jsonFile["Blocks"][blk]
                    timestamp = blockJson["MedianTime"]
                    secondOfDay = secondsIntoDay(timestamp)
                    day = dayNumber(timestamp)
                    year = yearNumber(timestamp)
                else:
                    end = True
                firstBlockOfDay = False
            firstBlockOfDay = True
            # These measure calls set the following for each loop:
            # minInnerCircumf
            # minLength
            # subUnitsMaxThickness
            # subUnitsTotalGapLengthWeight
            # Initially the following are set to these minima for now:
            # innerCircumf
            # length
            # The gist is that sizes of individual blocks will "percolate up" to all the higher level loops
            dayLoop.complete = True
            dayLoop.measure(daySpacingRatio)
            yearLoop.append(dayLoop)
        yearLoop.complete = True
        yearLoop.measure(yearSpacingRatio)
        centuryLoop.append(yearLoop)
        prevYear = year
    centuryLoop.measure(centurySpacingRatio)

    arcInnerCircumf = centuryLoop.innerCircumf
    yearsUnaccounted = 100 - len(centuryLoop.units)
    bigEndArcCircumfPerYear = centuryLoop.units[len(centuryLoop.units) - 1].length
    restOfCentury = yearsUnaccounted * bigEndArcCircumfPerYear
    centuryLoop.loopFraction = arcInnerCircumf / (arcInnerCircumf + restOfCentury)

    print("Pass Two Point One: Ramped high level measurements percolate down...")
    for y, yearLoop in enumerate(centuryLoop.units):
        for d, dayLoop in enumerate(yearLoop.units):
            avDayInnerCircumfRamped = 0.0
            for b, block in enumerate(dayLoop.units):
                avDayInnerCircumfRamped += dayLoop.innerCircumfRamped(b)
            avDayInnerCircumfRamped /= len(dayLoop.units)

            spacingPerLoop = avDayInnerCircumfRamped
            for b, block in enumerate(dayLoop.units):
                # Block length takes account of "ramped" circumf measurement of dayLoop at a particular block index
                # Here we are "percolating down" ramped day circumf to block length
                spacingBeforeThisBlock = spacingPerLoop * block.gapLengthWeightBefore / dayLoop.subUnitsTotalGapLengthWeight / 2.0
                spacingAfterThisBlock = spacingPerLoop * block.gapLengthWeightAfter / dayLoop.subUnitsTotalGapLengthWeight / 2.0
                blockLength = resultOrValue(block, "minLength") + spacingBeforeThisBlock + spacingAfterThisBlock
                block.length = max(block.length, blockLength)

    print("Pass Two Point Two: New low level measurements percolate up...")
    for y, yearLoop in enumerate(centuryLoop.units):
        for d, dayLoop in enumerate(yearLoop.units):
            dayLoop.measure(daySpacingRatio)
        yearLoop.measure(yearSpacingRatio)
    centuryLoop.measure(centurySpacingRatio)

    print("Pass Two Point Three: Ramped high level measurements percolate down...")
    for y, yearLoop in enumerate(centuryLoop.units):
        avYearInnerCircumfRamped = 0.0
        for d, dayLoop in enumerate(yearLoop.units):
            avYearInnerCircumfRamped += yearLoop.innerCircumfRamped(d)
        avYearInnerCircumfRamped /= len(yearLoop.units)
        spacingPerDay = avYearInnerCircumfRamped / len(yearLoop.units)

        for d, dayLoop in enumerate(yearLoop.units):
            # Block length takes account of "ramped" circumf measurement of yearLoop at a particular day index
            # Here we are "percolating down" ramped year circumf to day length
            dayLength = resultOrValue(dayLoop, "length") + spacingPerDay
            dayLoop.length = max(dayLoop.length, dayLength)

    print("Pass Two Point Four: Further low level measurements percolate up...")
    for y, yearLoop in enumerate(centuryLoop.units):
        # We've already percolated up the block level measurements.
        # To re-measure dayLoops would overwrite this. So we don't measure dayLoops this time around.
        yearLoop.measure(yearSpacingRatio)
    centuryLoop.measure(centurySpacingRatio)

    print("Third pass: Measure the positions...")
    for y, yearLoop in enumerate(centuryLoop.units):
        for d, dayLoop in enumerate(yearLoop.units):
            dayLoop.measurePositions()
        yearLoop.measurePositions()
    centuryLoop.measurePositions()

    print("Fourth pass, introduce transforms...")
    for y, yearLoop in enumerate(centuryLoop.units):
        for d, dayLoop in enumerate(yearLoop.units):
            for b, block in enumerate(dayLoop.units):

                # Special case for block 0, which is seperated by 6 days from block 1
                #if b==0 and d==0 and y==0:
                #    sixDays = 6 * dayLoop.length
                #    block.introducedTransforms.append(SpreadTranslateY(-sixDays, -sixDays))

                # Half block thickness so inside cylinder of dayLoop is smooth
                halfThickness = block.thickness / 2
                block.introducedTransforms.append(SpreadTranslateX(halfThickness, halfThickness))

                # Store some measurements of a "Base" so that block can render a base slab
                #block.baseLength = dayInnerRadius * 2.0 * math.pi * block["breadth"] * 1.01
                #block.baseWidth = yearRadiusAtDay * 2.0 * math.pi * dayLoop["breadth"] * 1.01

            # Give dayLoop a radius
            dayStartInnerRadius = dayLoop.startAttr("innerCircumf", True) / (2.0 * math.pi)
            dayEndInnerRadius = dayLoop.endAttr("innerCircumf", True) / (2.0 * math.pi)
            dayLoop.introducedTransforms.append(SpreadTranslateX(dayStartInnerRadius, dayEndInnerRadius))

            # Rotation for elements of dayLoop
            #dayLoop.introducedTransforms.append(SpreadRotateY(-90.0, -90.0))    # Rotate all so midnight is top
            startAngle = -dayLoop.startAngle    # Negated so clockwise
            endAngle = -dayLoop.endAngle
            dayLoop.introducedTransforms.append(SpreadRotateY(startAngle, endAngle))

        # Give yearLoop a radius
        yearStartRadius = yearLoop.startAttr("innerCircumf", True) / (2.0 * math.pi)
        yearEndRadius = yearLoop.endAttr("innerCircumf", True) / (2.0 * math.pi)
        yearLoop.introducedTransforms.append(SpreadTranslateX(yearStartRadius, yearEndRadius))

        # Rotation for elements of yearLoop
        yearLoop.introducedTransforms.append(SpreadRotateZ(0, yearLoop.loopFraction * 360.0))

    # Transforms introduced at centuryLoop
    arcInnerCircumf = centuryLoop.innerCircumf
    yearsUnaccounted = 100 - len(centuryLoop.units)
    bigEndArcCircumfPerYear = centuryLoop.units[len(centuryLoop.units) - 1].length
    restOfCentury = yearsUnaccounted * bigEndArcCircumfPerYear
    wholeCenturyRadius = (arcInnerCircumf + restOfCentury) / (2.0 * math.pi)
    centuryLoop.introducedTransforms.append(SpreadTranslateX(-wholeCenturyRadius, -wholeCenturyRadius))
    centuryLoop.introducedTransforms.append(SpreadRotateY(0, 360.0 * centuryLoop.loopFraction))
    centuryLoop.introducedTransforms.append(SpreadTranslateX(wholeCenturyRadius, wholeCenturyRadius))

    print( "Render..." )
    renderer = []
    centuryLoop.render(renderer, [])

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
