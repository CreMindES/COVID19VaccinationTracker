# COVID-19 Vaccination Tracker

A Twitter bot tracking COVID-19 Vacciation numbers in Hungary.

## Building

```bash
# just simply
go build
```

## Under the Hood

Simple 3-step prepation and tweet
- Fetch last tweet (needs bootstrapping the account) and parse last reported number
- Fetch latest vaccincation number from country specific source
- Fetch population from wikimedia
- Tweet status update
