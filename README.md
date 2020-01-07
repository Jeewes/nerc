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

An example of `nerc.yml` file:
```yaml
input: test_data/products.csv       # Defines the CSV input filepath
templates: "test_data/templates/"   # Defines the dirpath of the template files
output: output/                     # Defines the output dirpath
variables:                          # Defines template variables
  - key: ProductName                # Defines the variable key or "name"
    csvSourceCol: 6                 # CSV column no for product name
  - key: ProductPrice               
    csvSourceCol: 14
    type: price                     # Type of the variable. Price is rendered with two decimals.
  - key: ProductImage
    csvSourceCol: 21
  - key: VideoFile
    value: "video.avi"              # Hard coded value to be used
```
