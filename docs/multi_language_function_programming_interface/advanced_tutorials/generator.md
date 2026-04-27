# openYuanrong 生成器

Python 语言中，使用了 `yield` 关键字的函数被称为生成器（generator）。生成器函数是一种特殊的函数，可以在迭代过程中逐步产生结果，而不是一次性的返回所有的值。所有使用 `yr.instance` 或 `yr.invoke` 装饰的生成器称为 openYuanrong 生成器。

## 使用场景

生成器每次迭代只计算并返回一个值，无需一次性将所有值加载到内存中，可以有效地节省内存空间。常用场景包括：

- 无限序列：生成器可以用于生成无限序列，而不会耗尽内存。
- 数据流处理：生成器可以用于处理大量数据流，例如从文件中逐行读取数据或从网络流中接收数据。

## 使用示例

openYuanrong 生成器返回一个 `ObjectRefGenerator` 对象，通过该对象可调用 `_next` 及 `_anext` 方法，也可以迭代遍历。

遍历操作可能返回以下异常：

| **类型** | **描述** |
| -----------   | ------------------   |
| RuntimeError  | 获取 yield 返回值失败。 |
| YRInvokeError | 函数执行错误。          |

:::{note}

异步生成器只能在有状态函数中实现，无状态函数不支持异步。

:::

### 无状态函数生成器

```python
import yr

yr.init()

@yr.invoke
def countdown(n):
    while n > 0:
        yield n
        n -= 1

n = 5
# 使用 next 接口
generator = countdown.invoke(n)
print(yr.get(next(generator)))

# 使用 for 循环
for ref in generator:
    print(yr.get(ref))

yr.finalize()
```

### 有状态函数生成器

```python
import yr

yr.init()

@yr.instance
class StatefulFunc():
    def __init__(self, number):
        self.sum = number

    def countdown(self):
        n = self.sum
        while n > 0:
            yield n
            n -= 1

n = 5
instance = StatefulFunc.invoke(n)

# 使用 next 接口
generator = instance.countdown.invoke()
print(yr.get(next(generator)))

# 使用 for 循环
for ref in generator:
    print(yr.get(ref))

instance.terminate()
yr.finalize()
```

### 有状态函数异步生成器

```python
import asyncio
import time
import yr

yr.init()

@yr.instance
class StatefulFunc():
    async def gen(self, n):
        for i in range(n):
            await asyncio.sleep(1)
            yield i

instance = StatefulFunc.invoke()

async def async_gen(n):
    gen = instance.gen.invoke(n)
    async for ref in gen:
        res = await ref
        print(res)

async def multi_gen():
    n = 5
    await asyncio.gather(async_gen(n), async_gen(n), async_gen(n), async_gen(n), async_gen(n))

now = time.time()
asyncio.run(multi_gen())
cost = time.time() - now
# 并发执行 5 个 generator 函数，每个函数用时 5s, 总用时小于 6s
print(cost)

instance.terminate()
yr.finalize()
```
