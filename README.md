# NERC - nexrender configurer

Tool for filling nexrender config templates with CSV data.

**Note:** This is wery much work in progress...

## Getting started

1. Download `nerc` executable from releases
2. Configure `nerc.yml`
3. Run `nerc`
    - Run `nerc -h` to see help
4. You should now have bunch of nexrender configs in a new `output` directory

## Configuring nerc.yml

An example:
```yaml
input: test_data/products.csv
templates: "test_data/templates/"
output: output/
variables:
  ProductName: 6          # CSV column no for product name
  ProductPrice: 14        # CSV column no for product price
  ProductImage: 21        # CSV column no for product image
  VideoFile: "video.avi"  # Hard coded string value
```
