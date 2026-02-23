from itertools import cycle
import string
def f(word):
    '''
    Recall that if c is an ascii character then ord(c) returns its ascii code.
    Will be tested on nonempty strings of lowercase letters only.

    >>> f('x')
    The longest substring of consecutive letters has a length of 1.
    The leftmost such substring is x.
    >>> f('xy')
    The longest substring of consecutive letters has a length of 2.
    The leftmost such substring is xy.
    >>> f('ababcuvwaba')
    The longest substring of consecutive letters has a length of 3.
    The leftmost such substring is abc.
    >>> f('abbcedffghiefghiaaabbcdefgg')
    The longest substring of consecutive letters has a length of 6.
    The leftmost such substring is bcdefg.
    >>> f('abcabccdefcdefghacdef')
    The longest substring of consecutive letters has a length of 6.
    The leftmost such substring is cdefgh.
    '''
    #这道题只要递增的 压根不需要考虑z之后a那种 也压根不需要考虑cycle的事情
    
    max = "" 
    if word:
        max = word[0]
        cur = word[0]
        for i in range(1,len(word)):
            if ord(word[i]) == ord(word[i - 1])+1:
                cur += word[i]
            else:
                #这个判断不能写在这里，因为如果遍历到最后是一整个序列的话，这样是判断不到的
                #if len(max) < len(cur):
                #    max = cur
                cur = word[i]
            if len(max) < len(cur):
                max = cur
        print(f"The longest substring of consecutive letters has a length of {len(max)}.")
        print(f"The leftmost such substring is {max}.")
    #sample answer
    """
    longest = ""
    if word:
        longest = sub_string = word[0]
        for i in range(1, len(word)):
            if ord(word[i - 1]) + 1 == ord(word[i]):
                sub_string += word[i]
            else:
                sub_string = word[i]
            # 如果当前的长度，小于正在计算的长度
            if len(longest) < len(sub_string):
                longest = sub_string
    print(f"The longest substring of consecutive letters has a length of {len(longest)}.")
    print(f"The leftmost such substring is {longest}.")
    """
if __name__ == '__main__':
    import doctest

    doctest.testmod()
