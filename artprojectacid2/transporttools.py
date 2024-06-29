import base64
from array import array

def floatArrayToString(arr):
        float_array = array('f',arr)
        byts = float_array.tobytes()
        b64 = base64.b64encode(byts)
        return b64.decode('ascii')
