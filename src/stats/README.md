### 统计服务器

用于统计玩家行为，并周期性刷入统计数据库。 刷入的内容包括：     
1. 一个时间段的操作summary。（如：一天）       
2. 刷入时玩家基本数据(当前的资源，建筑情况等）的snapshot二者。       

持续性生产的计算方法,产量可表达为     

f(T) = Max(S + (T-T0) x RATE, Capacity) (S是T0时刻的储量, Capacity是容量)      

一旦T1时刻要使用产出资源，那么:     
S = f(T1) - COST      
NP  = (T1-T0) x RATE , NP是这段时间的自然产出量       

GS 通知stats服务器NP值, stats服务器纪录这个值      

stats服务器最终统计的产出值，等于 Sum(NP) + f(Tn)  (Tn是统计时刻的时间)       
