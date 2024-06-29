from spiraltools import *
import json

# HANDEDNESS
# This python code produces files for two target render platforms: Unreal Engine and Three.js.
# Three.js is a web platform and so needs a compact file to be served - hence quaternions
# are favoured for export.
# We use a python quaternion library which presumably has a handedness we should be aware of.
# The python code here will have a handedness to match this quaternion library.
# A python test shows that in the quaternion library, (1,0,0) rotated 90' around (0,1,0)
# gives (0,0,-1). This is compatible with
# The following diagrams disagree with what I consider to be a clockwise rotation
# (They seem to mean clockwise when looking down the NEGATIVE axis. Doh!)
# https://www.evl.uic.edu/ralph/508S98/coordinates.html
# The above test is compatible with the diagram's Left handed system,
# with rotation being COUNTER-CLOCKWISE when looking towards the +ve end of the rotation axis.
# The above test is ALSO compatible with the diagram's Right handed system,
# with rotation being CLOCKWISE when looking towards the +ve end of the rotation axis!
# For compatibility with the original concept (Unreal Engine based), we consider the
# "floor" to be the x-y plane. X is line of sight and Y to the right, with Z up.
# This is therefore a LEFT-HANDED system, with rotations COUNTER-CLOCKWISE looking +ve down the axis of rotation.

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
        instanceTransform.append(ScaleZ(self.minLength))
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
            colouredSlab = Cube(self.baseR, self.baseG, self.baseB, 1,0, False, False)     # Orange
            # Apply an extra transform to base of slab
            slabTransform.append(TranslateX(-self.thickness * 0.505))
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

    daySpacingRatio = 1.1
    yearSpacingRatio = 1.3
    centurySpacingRatio = 1.0

    print( "Opening source data file...")
    fi1 = open("Input\\acidblocks.json")
    jsonFile = json.load(fi1)
    fi1.close()

    print( "First pass: populate, and measure to percolate up...")
    centuryLoop = Loop()
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
                # Length (circumferential) is the shortest measurement for small blocks
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
                red = blockJson["ColourByte0"]
                green = blockJson["ColourByte1"]
                blue = blockJson["ColourByte2"]
                # Bitcoin orange
                baseR = 255.0 / 256.0
                baseG = 153.0 / 255.0
                baseB = 0.0
                includeBase = False
                # Assume not last of day and year until we find otherwise
                block = Block(length, width, thickness, red, green, blue, includeBase, baseR, baseG, baseB)
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

            prevD = d
            # These measure calls set the following for each loop:
            # minInnerCircumf
            # minLength
            # subUnitsMaxThickness
            # Initially the following are set to these minima for now:
            # innerCircumf
            # length
            # The gist is that sizes of individual blocks will "percolate up" to all the higher level loops
            dayLoop.complete = True
            dayLoop.measure(daySpacingRatio)
            yearLoop.append(dayLoop)

        prevY = y
        yearLoop.complete = True
        yearLoop.measure(yearSpacingRatio)
        centuryLoop.append(yearLoop)
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

            spacingPerBlock = avDayInnerCircumfRamped / len(dayLoop.units)
            for b, block in enumerate(dayLoop.units):
                # Block length takes account of "ramped" circumf measurement of dayLoop at a particular block index
                # Here we are "percolating down" ramped day circumf to block length
                blockLength = resultOrValue(block, "minLength") + spacingPerBlock
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
        spacingPerBlock = avYearInnerCircumfRamped / len(yearLoop.units)

        for d, dayLoop in enumerate(yearLoop.units):
            # Block length takes account of "ramped" circumf measurement of yearLoop at a particular day index
            # Here we are "percolating down" ramped year circumf to day length
            dayLength = resultOrValue(dayLoop, "length") + spacingPerBlock
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

                # Half block thickness so inside cylinder of dayLoop is smooth
                halfThickness = block.thickness / 2
                # THIS IS THE FIRST POINT IN THE CODE WHERE WE CARE ABOUT ORIENTATION
                # OF X,Y,Z AXES AND CLOCKWISE/COUNTERCLOCKWISE
                # By Decree:
                # X is a "floor" axis going from the centre of day 0's day loop towards
                # the outside of the day loop.
                block.introducedTransforms.append(SpreadTranslateX(halfThickness, halfThickness))

                # Store some measurements of a "Base" so that block can render a base slab
                block.baseLength = block.length * 1.01
                block.baseWidth = dayLoop.length * 1.01

            # Give dayLoop a radius
            dayStartInnerRadius = dayLoop.startAttr("innerCircumf", True) / (2.0 * math.pi)
            dayEndInnerRadius = dayLoop.endAttr("innerCircumf", True) / (2.0 * math.pi)
            dayLoop.introducedTransforms.append(SpreadTranslateX(dayStartInnerRadius, dayEndInnerRadius))

            # Rotation for elements of dayLoop
            # By decree:
            # Y is a "floor" axis that is the axis through the centre of day 0's day loop.
            # Positive rotations around Y, when looking in a +ve Y direction, are COUNTER-CLOCKWISE
            # (as per Left Handed System)
            dayLoop.introducedTransforms.append(SpreadRotateY(0, 360.0))

        # Give yearLoop a radius
        yearStartRadius = yearLoop.startAttr("innerCircumf", True) / (2.0 * math.pi)
        yearEndRadius = yearLoop.endAttr("innerCircumf", True) / (2.0 * math.pi)
        yearLoop.introducedTransforms.append(SpreadTranslateX(yearStartRadius, yearEndRadius))

        # Rotation for elements of yearLoop
        # By decree:
        # The z axis is vertical (positive up), and is the axis for year loops
        # Positive rotations, when looking towards the +ve end of the rotation axis,
        # are COUNTER-CLOCKWISE, as per left handed system
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

    print( "Compose transforms..." )
    for instance in renderer:
        instance.composeTransform()

    print( "Save..." )
    fo = open("Output\\renderspecquat.json", 'w')
    json.dump(renderer, fo, default=vars, indent=2)

    print( "Save Subset...")
    # Throw away all except first 100000 blocks
    subrenderer = []
    for b, block in enumerate(renderer):
        if b < 300000:
            subrenderer.append(block)

    fo = open("Output\\renderspecsub.json", 'w')
    json.dump(subrenderer, fo, default=vars, indent=2)

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
                red = blockJson["ColourByte0"]
                green = blockJson["ColourByte1"]
                blue = blockJson["ColourByte2"]
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
