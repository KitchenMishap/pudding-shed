import math
from pyquaternion import Quaternion
import numpy
from transporttools import *

def real_first_test():
    print("real_first_test")
    quat = Quaternion()
    # Is it the identity?
    quatinv = quat.inverse;
    assert(quat==quatinv)
    # Is the first element the real part?
    assert(quat.elements[0]==1.)

def right_handed_test():
    print("right_handed_test()")
    # Check that x axis rotated by 90 degrees around y axis is negative z axis
    numpyx = numpy.array([1.,0.,0.])
    numpyy = numpy.array([0.,1.,0.])
    numpyz = numpy.array([0.,0.,1.])
    roty = Quaternion(axis=numpyy, angle=math.pi/2)
    v = roty.rotate(numpyx)
    print(v)
    assert(math.fabs(v[0]) < 0.0001)
    assert(math.fabs(v[1]) < 0.0001)
    assert(math.fabs(v[2] - -1.) < 0.0001)
    # https://www.evl.uic.edu/ralph/508S98/coordinates.html
    # https://techarthub.com/a-practical-guide-to-unreal-engines-coordinate-system/
    # Unreal uses left handed co-ordinates, Z up.
    # So X is forward, Y is right (on the floor) and Z is up (above the floor).
    # Also, left handed means +ve angles are clockwise.
    # So Rotating an x vector by y axis clockwise gives z vector (NOT what we find above)
    # THEREFORE, these quaternions are NOT compatible with Unreal!
    # WE WILL NEED TO INVERT THE X AXIS BEFORE WE USE THESE QUATERNIONS

def quat_chain_order_test():
    numpyx = numpy.array([1.,0.,0.])
    numpyy = numpy.array([0.,1.,0.])
    numpyz = numpy.array([0.,0.,1.])
    rotx = Quaternion(axis=numpyx, angle=math.pi/2.)
    roty = Quaternion(axis=numpyy, angle=math.pi/2.)
    rotz = Quaternion(axis=numpyz, angle=math.pi/2.)
    ythenx = rotx * roty
    result = ythenx.rotate(numpyx)
    # X axis, rotated by y then x axes, should give y axis
    assert(math.fabs(result[0]) < 0.0001)
    assert(math.fabs(result[1] - 1.) < 0.0001)
    assert(math.fabs(result[2]) < 0.0001)

def twos_test():
    twos = [2.,2.,2.]
    assert(floatArrayToString(twos)=="AAAAQAAAAEAAAABA")

print("Running tests...")
real_first_test()
right_handed_test()
quat_chain_order_test()
twos_test()
print("Tests have been run")
