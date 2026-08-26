# Vendor Specific Options

DHCP from its inception was meant to be extensible with little effort. One of
the largest parts are vendor specific options provided to clients. With PG DHCP,
you can define custom options to be encapsulated within the vendor specific
information option (option 43).

## Declaring Option Types

Every option used in an option 43 segment must be define. The definition simply
needs the option code and data type.

```
decloption test-opt
    code 18
    type string
end
```

The above example declares a new vendor option named `test-opt`. All vendor names
must be unique within the configuration but they can be any string value. The
names are only used internally and not sent to the client.

The option code is 18 with a type string. All vendor options are type checked
just like normal options. Using the option is the same as using any other option.

```
network network1
    subnet 10.0.2.0/24
        range 10.0.2.10 10.0.2.200
        option router 10.0.2.1
        option test-opt "test"
        option vendor-options true
    end
end
```

The custom option is defined in the subnet with the value "test".

**NOTE:** The option `vendor-options` must be set to true in a subnet where you
want the vendor information to be sent. If `vendor-options` is not set or is
false, vendor options will NOT be sent. Also, make sure all vendor options are
defined BEFORE declaring `vendor-options` to true.

## Vendor Options Codes

Vendor codes can be any number between 1 and 254 inclusive. You will need to
consult the vendor documentation for which code any type to use.

## Vendor Option Types

The available types for vendor options are:

- `bool` - Boolean, true/false
- `address` - Single IPv4 address
- `address-list` - List of IPv4 addresses
- `string` - String
- `int8` - 8 bit number (0 - 255)
- `int16` - 16 bit number (0 - 65535)
- `int32` - 32 bit number (0 - 2147483647)
