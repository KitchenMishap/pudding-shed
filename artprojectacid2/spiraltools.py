import math

def resultOrValue(object, attrName):
    attr = getattr(object, attrName, None)
    if attr is None and isinstance(object, dict):
        attr = object[attrName]
    if callable(attr):
        return attr()
    return attr

#region Asset
# An asset is a drawable object
# It has a name such as "Cube", and five numbers known as r,g,b,a,h

class Asset:
    def __init__(self, name, r,g,b,a,h):
        self.name = name
        self.r = r
        self.g = g
        self.b = b
        self.a = a
        self.h = h

def Cube(r,g,b,a,h):
    return Asset("Cube", r,g,b,a,h)

def Sphere(r,g,b,a,h):
    return Asset("Sphere", r,g,b,a,h)
#endregion

#region Transform
# A transform is a list of TransformPrimitives
# A TransformPrimitive has one of nine names, and an amount

TransformNames = ["ScaleX", "ScaleY", "ScaleZ",
                  "TranslateX", "TranslateY", "TranslateZ",
                  "RotateX", "RotateY", "RotateZ"]

class TransformPrimitive:
    def __init__(self, name, amount):
        self.name = name
        self.amount = amount

def ScaleX(scale):
    return TransformPrimitive("ScaleX", scale)
def ScaleY(scale):
    return TransformPrimitive("ScaleY", scale)
def ScaleZ(scale):
    return TransformPrimitive("ScaleZ", scale)
def TranslateX(distance):
    return TransformPrimitive("TranslateX", distance)
def TranslateY(distance):
    return TransformPrimitive("TranslateY", distance)
def TranslateZ(distance):
    return TransformPrimitive("TranslateZ", distance)
def RotateX(angle):
    return TransformPrimitive("RotateX", angle)
def RotateY(angle):
    return TransformPrimitive("RotateY", angle)
def RotateZ(angle):
    return TransformPrimitive("RotateZ", angle)

class SpreadableTransform:
    def __init__(self, name, startAmount, endAmount):
        self.name = name
        self.start = startAmount
        self.end = endAmount

    def InterpolateTransform(self, position):
        amount = self.start + position * (self.end - self.start)
        return TransformPrimitive(self.name, amount)

    def SpreadInterpolate(self, startPosition, endPosition):
        startAmount = self.InterpolateTransform(startPosition).amount
        endAmount = self.InterpolateTransform(endPosition).amount
        return SpreadableTransform(self.name, startAmount, endAmount)

def SpreadTranslateX(startDistance, endDistance):
    return SpreadableTransform("TranslateX", startDistance, endDistance)
def SpreadTranslateY(startDistance, endDistance):
    return SpreadableTransform("TranslateY", startDistance, endDistance)
def SpreadTranslateZ(startDistance, endDistance):
    return SpreadableTransform("TranslateZ", startDistance, endDistance)
def SpreadRotateX(startAngle, endAngle):
    return SpreadableTransform("RotateX", startAngle, endAngle)
def SpreadRotateY(startAngle, endAngle):
    return SpreadableTransform("RotateY", startAngle, endAngle)
def SpreadRotateZ(startAngle, endAngle):
    return SpreadableTransform("RotateZ", startAngle, endAngle)

#endregion

#region Instance
class Instance(dict):
    def __init__(self, asset, transform):
        super().__init__()
        self["asset"] = asset
        self["transform"] = transform

    def render(self, renderer, transform):
        self["transform"] = self["transform"] + transform
        renderer.append(self)
#endregion

#region Loop
class Loop(dict):
    def __init__(self):
        super().__init__()
        self.units = []                 # units are Volumes or other Loops
        self.introducedTransforms = []

    def append(self, unit):
        self.units.append(unit)

    def measure(self, spacingRatio):
        # Calculates various items for self, and position/breadth for each contained unit
        self.unspacedCircumf = 0
        self.maxWidth = 0
        self.maxThickness = 0
        for unit in self.units:
            # length can be a value or a function to be called
            length = resultOrValue(unit, "length" )
            self.unspacedCircumf += length

            width = resultOrValue(unit, "width")
            self.maxWidth = max(self.maxWidth, width)

            thickness = resultOrValue(unit, "thickness")
            self.maxThickness = max(self.maxThickness, thickness)

        # Have to do this in a new loop, to use self.unspacedCircumf
        runningTotal = 0
        for unit in self.units:
            length = resultOrValue(unit, "length")
            # unit.position is a number between 0 an 1
            unit["position"] = (runningTotal + length/2) / self.unspacedCircumf
            runningTotal += length
            # unit.breadth is also a number between 0 and 1 on the same scale
            unit["breadth"] = length / self.unspacedCircumf

        self.spacingRatio = spacingRatio

    def innerCircumf(self):
        return self.unspacedCircumf * self.spacingRatio

    def innerRadius(self):
        return self.innerCircumf() / (2 * math.pi)

    def outerRadius(self):
        return self.innerCircumf() / (2*math.pi) + self.maxThickness

    def length(self):
        return self.maxWidth

    def width(self):
        return 2 * self.outerRadius()

    def thickness(self):
        # A loop is like a disc/cylinder, so width = thickness
        return 2 * self.outerRadius()

    # A ramped version of an attribute is guaranteed to be greater than the attribute itself,
    # and is continuous (no sudden jumps). If the attribute is greater in the previous loop,
    # the first third ramps down from that value. If the attribute is greater in the next loop,
    # the last third ramps up to that value.
    def rampedAttr(self, attrName, index, prevLoop, nextLoop):
        count = len(self.units)
        currAttr = getattr(self, attrName)
        if index < count / 3 and not (prevLoop is None):
            prevAttr = getattr(prevLoop, attrName)
            ramp = index / (count / 3)
            val = (1.0 - ramp) * prevAttr + ramp * currAttr
        elif index > 2 * count / 3 and not (nextLoop is None):
            nextAttr = getattr(nextLoop, attrName)
            ramp = (index - 2 * count / 3) / (count / 3)
            val = (1.0 - ramp) * currAttr + ramp * nextAttr
        else:
            val = currAttr
        return val

    def render(self, renderer, delegatedTransforms):
        innerRadius = self.innerRadius()
        for unit in self.units:
            subUnitTransforms = []
            position = unit["position"]
            breadth = unit["breadth"]
            startPosition = position - breadth/2
            endPosition = position + breadth/2

            # Introduce some transforms to sub-units, spreading according to position spread
            for introduced in self.introducedTransforms:
                scaledSpreadable = introduced.SpreadInterpolate(startPosition, endPosition)
                subUnitTransforms.append(scaledSpreadable)

            # Apply delegatedTransforms to sub-units, spreading according to position spread
            for delegated in delegatedTransforms:
                scaledSpreadable = delegated.SpreadInterpolate(startPosition, endPosition)
                subUnitTransforms.append(scaledSpreadable)

            unit.render(renderer, subUnitTransforms)
#endregion
