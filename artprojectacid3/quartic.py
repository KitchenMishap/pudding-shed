import numpy as np
import math

def solve_for_c(K, Y):
    """
    Solves the linear system KC = Y for C.

    Parameters:
    -----------
    K : numpy.ndarray
        A 5x5 coefficient matrix
    Y : numpy.ndarray
        A column vector (5x1 or 1D array of length 5)

    Returns:
    --------
    C : numpy.ndarray
        The solution vector

    Example:
    --------
    >>> K = np.array([[2, 1, 0, 0, 0],
    ...               [1, 3, 1, 0, 0],
    ...               [0, 1, 4, 1, 0],
    ...               [0, 0, 1, 3, 1],
    ...               [0, 0, 0, 1, 2]])
    >>> Y = np.array([1, 2, 3, 4, 5])
    >>> C = solve_for_c(K, Y)
    >>> print(C)
    """
    # Ensure Y is a 1D array
    Y = np.asarray(Y).flatten()

    # Solve the linear system using numpy's linear solver
    C = np.linalg.solve(K, Y)                       # This was Claude.ai's first choice
    #C = np.linalg.lstsq(K, Y, rcond=None)[0]       # Useful if K might be singular

    return C

def make_quartic_curve(x1, y1, g1, x2, y2, g2, x3, y3):
    # (x1, y1) is a point through the quartic with gradient g1
    # (x2, y2) is a point through the quartic with gradient g2
    # (x3, y3) is a point on the quartic with unknown gradient
    K = np.array([[math.pow(x1, 4), math.pow(x1, 3), math.pow(x1, 2), x1, 1.0],
                  [math.pow(x2, 4), math.pow(x2, 3), math.pow(x2, 2), x2, 1.0],
                  [math.pow(x3, 4), math.pow(x3, 3), math.pow(x3, 2), x3, 1.0],
                  [4.0 * math.pow(x1, 3), 3.0 * math.pow(x1, 2), 2.0 * x1, 1.0, 0.0],
                  [4.0 * math.pow(x2, 3), 3.0 * math.pow(x2, 2), 2.0 * x2, 1.0, 0.0]], dtype=float)
    Y = np.array([y1, y2, y3, g1, g2], dtype = float)
    C = solve_for_c(K, Y)
    # C is the column vector [a,b,c,d,e] of coefficients describing the
    # quartic equation y = ax^4 + bx^3 + cx^2 + dx + e
    return C

def quartic_curve(x, C):
    return C[0] * math.pow(x, 4.0) + C[1] * math.pow(x, 3.0) + C[2] * math.pow(x, 2.0) + C[3] * x + C[4]

def quartic_above_points(points, C):
    # Checks whether the given quartic C has y >= points[i][1] for i = all except left and right points
    length = len(points)
    for i in range(1, length-1):
        xi = points[i][0]
        yi = points[i][1]
        y = quartic_curve(xi, C)
        if y < yi:
            return False
    return True

def make_quartic_dip(points, iterations):
    length = len(points)
    # Left point, assumed to be gradient zero
    x1 = points[0][0]
    y1 = points[0][1]
    g1 = 0.0
    # Right point, assumed to be gradient zero
    x2 = points[length - 1][0]
    y2 = points[length - 1][1]
    g2 = 0.0

    # Mid point, which we vary up and down between a bottom and top
    xmid = (x1+x2)/2.0
    ybottom = 0.0
    ytop = (y1+y2)/2.0
    # We assume that this curve will succeed, so it is the best so far
    Cbest = make_quartic_curve(x1,y1,0.0,x2,y2,0.0,xmid,ytop)

    for i in range(iterations):
        ytry = (ybottom + ytop)/2.0
        C = make_quartic_curve(x1,y1,0.0,x2,y2,0.0,xmid,ytry)
        success = quartic_above_points(points, C)
        if success:
            # A quartic including the point ytry was successfully above all the points.
            Cbest = C.copy()
            # Try lowering ytop to see if we can do even better
            ytop = (ybottom+ytop) / 2.0
        else:
            # A quartic including the point ytry was not above all the points.
            # Try raising ybottom to see if we can get back into the success zone
            ybottom = (ybottom + ytop)/2.0
    return Cbest
