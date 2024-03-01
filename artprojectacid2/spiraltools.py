import math

def resultOrValue(object, attrName):
    if type(object) is dict:
        attr = object[attrName]
    else:
        attr = getattr(object, attrName)
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
#endregion
#region Instance
class Instance:
    def __init__(self, asset, transforms):
        self.asset = asset
        self.transform = transforms
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
            # length can be an value or a function to be called
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
#endregion
#region DayLoop
class DayLoop(Loop):
    def render(self, renderer, transforms):
        innerRadius = self.innerRadius()
        for unit in self.units:
            position = unit["position"]
            overallTransforms = []
            overallTransforms.append(ScaleX(unit["thickness"]))
            overallTransforms.append(ScaleY(unit["width"]))
            overallTransforms.append(ScaleZ(unit["length"]))
            radius = innerRadius + unit["thickness"] / 2
            overallTransforms.append(TranslateX(radius))
            dayLength = self.length()
            shift = position * dayLength - dayLength / 2
            overallTransforms.append(TranslateY(shift))
            angle = position * 360
            overallTransforms.append(RotateY(angle))

            overallTransforms += transforms

            renderer.append( Instance( Cube(0,1,0,1,0), overallTransforms ) )
