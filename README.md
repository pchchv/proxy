<div align="center">

## **A small cached HTTP proxy**

</div>

<div align="center">

```
           network            proxy             cache
         .---------.       .---------.       .---------.
    <----|- - < - -|---<---|- - < - -|---<---|- < -.   |
you ---->|- - > - -|--->---|- -,- > -|--->---|- > -|   |
         |         |       |   |(*)  |       |     |   |
         |    ,-< -|---<---|< -'     |       |     |   |
         |    , ,->|--->---|- - > - -|--->---|- > -'   |
         `----+-+--´       `---------´       `---------´
              ' '
              '_'
            website
```

</div>

###### (*) When the data is not in the cache, the website will be requested and is directly stored in the cache.

###### Where "network" may be anything (LAN/WAN/...).

#
