# ord(c) returns the encoding of character c.
# chr(e) returns the character encoded by e.

from itertools import cycle
import string
def rectangle(width, height):
    '''
    Displays a rectangle by outputting lowercase letters, starting with a,
    in a "snakelike" manner, from left to right, then from right to left,
    then from left to right, then from right to left, wrapping around when z is reached.
    
    >>> rectangle(1, 1)
    a
    >>> rectangle(2, 3)
    ab
    dc
    ef
    >>> rectangle(3, 2)
    abc
    fed
    >>> rectangle(17, 4)
    abcdefghijklmnopq
    hgfedcbazyxwvutsr
    ijklmnopqrstuvwxy
    ponmlkjihgfedcbaz
    '''
    # REPLACE THE PREVIOUS LINE WITH YOUR CODE
    cyc = cycle(string.ascii_lowercase)
    for i in range(height):
        line = ""
        for j in range(width):
            line += next(cyc)
        if i % 2 != 0:
            line = line[::-1]
        print(line)
    return

#sample answer
"""
    start = ord('a')
    line = ""
    # 输出字母
    for i in range(width * height):
        line += chr(start + i % 26)

    # 输出结果
    for i in range(0, len(line), width):
        # i 的取值范围是0, width, 2 * width, 3 * width, 4 * width ..
        # 如果width 5, 则 i 的取值范围是 0, 5, 10, 15, 20, 25..
        line_no = i // width
        # 每行的输出内容
        each_line = line[i: i + width]
        if line_no % 2 == 0:
            print(each_line)
        else:
            print(each_line[::-1])
"""
if __name__ == '__main__':
    import doctest
    doctest.testmod()
