'''
Will be tested with height a strictly positive integer.
'''
from itertools import cycle
import string
def f(height):
    '''
    >>> f(1)
    0
    >>> f(2)
     0
    123
    >>> f(3)
      0
     123
    45678
    >>> f(4)
       0
      123
     45678
    9012345
    >>> f(5)
        0
       123
      45678
     9012345
    678901234
    >>> f(6)
         0
        123
       45678
      9012345
     678901234
    56789012345
    >>> f(20)
                       0
                      123
                     45678
                    9012345
                   678901234
                  56789012345
                 6789012345678
                901234567890123
               45678901234567890
              1234567890123456789
             012345678901234567890
            12345678901234567890123
           4567890123456789012345678
          901234567890123456789012345
         67890123456789012345678901234
        5678901234567890123456789012345
       678901234567890123456789012345678
      90123456789012345678901234567890123
     4567890123456789012345678901234567890
    123456789012345678901234567890123456789
    '''
    # Insert your code here
    cyc = cycle(string.digits)
    for i in range(height):
        line = ""
        space = " " * (height - i -1)
        for times in range(0,1+(i)*2):
            line += next(cyc)
        print(space+line)
    return
#sample answer
"""
    count = 0
    for row in range(height):
        print(" " * (height - row - 1), end="")
        for col in range( 2 * row + 1):
            print(count % 10, end="")
            count += 1
        print()
"""
if __name__ == '__main__':
    import doctest

    doctest.testmod()
