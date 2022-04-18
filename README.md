# apnic-asn-prefix-list
Parses ASN list from apnic and sorts by country and ASN

# Usage

```bash
$ ./apnic-asn-prefix-list
  -as string
        provide an ASN to print its IP prefixes
  -country string
        provide a country code to print all ASNs' IP prefixes in this country
```

For example lets try AS133938

```bash
$ ./apnic-asn-prefix-list -as AS133938
103.48.119.0/24
103.108.140.0/24
103.132.96.0/24
103.153.240.0/24
103.153.241.0/24
```

You can also try with two letter country code to get all prefixes for that country. For example:

```bash 
$ ./apnic-asn-prefix-list -country NZ
```

Output omitted for obvious reasons :P