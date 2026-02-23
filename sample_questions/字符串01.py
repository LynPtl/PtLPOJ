def remove_consecutive_duplicates(word):
    '''
    >>> remove_consecutive_duplicates('')
    ''
    >>> remove_consecutive_duplicates('a')
    'a'
    >>> remove_consecutive_duplicates('ab')
    'ab'
    >>> remove_consecutive_duplicates('aba')
    'aba'
    >>> remove_consecutive_duplicates('aaabbbbbaaa')
    'aba'
    >>> remove_consecutive_duplicates('abcaaabbbcccabc')
    'abcabcabc'
    >>> remove_consecutive_duplicates('aaabbbbbaaacaacdddd')
    'abacacd'
    '''
    # Insert your code here (the output is returned, not printed out)
    L = list(word)
    if word:
        cur = L[0]
        result = [L[0]]
        for i in range(1,len(word)):
            if L[i] == cur:
                pass
            else:
                cur = L[i]
                result.append(L[i])
        return "".join(result)
    else:
        return ""
    

#sample answer
"""
    # Insert your code here (the output is returned, not printed out
    # 双索引/双元素/双指针算法
    # 第一种方式
    # result = ""
    # if word:
    #     result = first = word[0]
    #     for second in word[1:]:
    #         if first == second:
    #             continue
    #         else:
    #             result += second
    #         first = second
    # return result
    result = ""
    if word:
        result = word[0]
        for i in range(1, len(word)):
            if word[i - 1] == word[i]:
                continue

            result += word[i]
    return result
"""

if __name__ == '__main__':
    import doctest

    doctest.testmod()
