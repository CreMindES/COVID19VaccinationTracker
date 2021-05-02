# COVID-19 Vaccination Tracker

A Twitter bot tracking COVID-19 Vacciation numbers in Hungary.

<p align=center>
  <img height=120 src=https://user-images.githubusercontent.com/5306361/116821253-71964c80-ab79-11eb-8d9d-a634799e1297.png>
</p>

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
