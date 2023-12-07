# Custom Types

| yml             | c                                        |
| --------------- | ---------------------------------------- |
| `bool`          | `WGPUBool`                               |
| `string`        | `char const *`                           |
| `uint16`        | `uint16_t`                               |
| `uint32`        | `uint32_t`                               |
| `uint64`        | `uint64_t`                               |
| `usize`         | `size_t`                                 |
| `int16`         | `int16_t`                                |
| `float32`       | `float`                                  |
| `float64`       | `double`                                 |
| `c_void`        | `void`                                   |
| (maybe) `array` | `T *` / `T const *` + `size_t` for count |

# Complex types

- `enum.*`
- `struct.*`
- `callback.*`
- `object.*`

# Struct type

|                 |                                                                                 |
| --------------- | ------------------------------------------------------------------------------- |
| `base_in`       | base struct that can be extended, and used to feed in data                      |
| `base_out`      | base struct that can be extended, and used to get result                        |
| `extension_in`  | extension struct that can be chained in a base struct, and used to feed in data |
| `extension_out` | extension struct that can be chained in a base struct, and used to get result   |
| `standalone`    | standalone structs (WGPULimits, WGPUColor, etc)                                 |

