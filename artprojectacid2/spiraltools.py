import math

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
#endregion
#region Instance
class Instance:
    def __init__(self, asset, transform):
        self.asset = asset
        self.transform = transform
#endregion
#region Loop
class Loop:
    def __init__(self):
        self.units = []             # units are Volumes or other Loops

    def append(self, unit):
        self.units.append(unit)

    def process(self, spacingRatio):
        # Calculates various items for self and position for each contained unit
        self.unspacedCircumf = 0
        self.maxWidth = 0
        self.maxThickness = 0
        for unit in self.units:
            # Prefer length attribute, fall back to Length() call
            length = getattr(unit, "length", unit.Length())
            self.unspacedCircumf += length
            # Prefer width attribute, fall back to Width() call
            width = getattr(unit, "width", unit.Width())
            self.maxWidth = max(self.maxWidth, width)
            # Prefer thickness attribute, fall back to Thickness() call
            thickness = getattr(unit, "thickness", unit.Thickness())
            self.maxThickness = max(self.maxThickness, unit.Thickness())
        # Have to do this in a new loop, to use self.unspacedCircumf
        runningTotal = 0
        for unit in self.units:
            # Prefer length attribute, fall back to Length() call
            length = getattr(unit, "length", unit.Length())
            # unit.position is a number between 0 an 1
            unit.position = (runningTotal + length/2) / self.unspacedCircumf
            runningTotal += length
        self.spacingRatio = spacingRatio

    def innerCircumf(self):
        return self.unspacedCircumf * self.spacingRatio

    def outerRadius(self):
        return self.innerCircumf() / (2*math.pi) + self.maxThickness

    def Length(self):
        return self.maxWidth

    def Width(self):
        return 2 * self.outerRadius()

    def Thickness(self):
        # A loop is like a disc/cylinder, so width = thickness
        return 2 * self.outerRadius()
#endregion
