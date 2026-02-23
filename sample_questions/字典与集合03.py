'''
Will be tested with year between 1913 and 2013.
You might find the reader() function of the csv module useful,
but you can also use the split() method of the str class.
'''
import csv
from collections import defaultdict

def f(year):
    '''
    >>> f(1914)
    In 1914, maximum inflation was: 2.0
    It was achieved in the following months: Aug
    >>> f(1922)
    In 1922, maximum inflation was: 0.6
    It was achieved in the following months: Jul, Oct, Nov, Dec
    >>> f(1995)
    In 1995, maximum inflation was: 0.4
    It was achieved in the following months: Jan, Feb
    >>> f(2013)
    In 2013, maximum inflation was: 0.82
    It was achieved in the following months: Feb
    '''
    months = 'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'
    # Insert your code here
    from collections import defaultdict
    counter = defaultdict(list)
    # 必须在同一个目录
    with open('cpiai.csv') as csvfile:
        reader = csv.DictReader(csvfile)
        # 以字典的形式
        for row in reader:
            # 读取csv中的date,并且根据字符串来分割
            y,m,d = row['Date'].split("-")
            # 查找等于目标的年份
            if y == str(year):
                counter[float(row['Inflation'])].append(int(m))
    # 求最大值： 一定要先判断是否为空
    if counter:
        # 求出最大的通货膨胀率
        max_flation = max(counter)
        # 获取对应的月份:[1, 2, 1, 2, 3, 2, 9, 11, 9, 10, 12]
        mon_values = sorted(set(counter[max_flation]))
        # 通过对应的月份: Jan, Feb, Mar, Apr, May
        values =(months[m - 1] for m in mon_values)
        # 直接输出结果
        print(f"In {year}, maximum inflation was: {max_flation}")
        print(f"It was achieved in the following months: {', '.join(values)}")

if __name__ == '__main__':
    import doctest
    doctest.testmod()
