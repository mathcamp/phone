# phone

Package phone is a package to parse phone numbers from strings. It figures out the country and the other details for a given phone number and supports 156 countries as of now. The country details are given in the raw/country-phones.json file. This file also has phone number formatting for various countries and can be used on the client as well. 

It also supports checking for test and toll free phone numbers within the US. 

## Usage

```
p, err := phone.ParseNumber("+11235556677") // For details about the parsed structure look at the godocs

```
