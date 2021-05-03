# COVID-19 Vaccination Tracker

A Twitter bot tracking COVID-19 Vacciation numbers in Hungary.

> Nothing fancy, just a simple Twitter bot running as an Azure Function.

<p align=center><a href="https://twitter.com/HUVaccineCount">
  <img height=120 src=https://user-images.githubusercontent.com/5306361/116821253-71964c80-ab79-11eb-8d9d-a634799e1297.png>
</a></p>

## Building

```bash
# just simply
go build -o hucovidtracker
```

## Deployment

Deploy as an Azure Function.

1. Create an account
2. Create a resource group
    ```bash
    deploy_region="westeurope"                                                                                    
    deploy_group="go-custom-function"                                                                            
    az group create -l $deploy_region -n $deploy_group 
    ```
3. Build
4. Create resources for Azure Function from deployment ARM (Azure Resource Management) template.
    ```bash
    az deployment group create --resource-group $deploy_group --template-file azuredeploy.json
    ```
5. Set API keys and secrets @ portal.azure.com.
   
   `$deploy_group > azure function > Settings|Configuration > New application settings x 4`
6. Deploy
    ```bash
    func azure functionapp publish `[function name/ID]` --no-build --force
    ```

## Under the Hood

Simple 3-step prepation and tweet
- Fetch last tweet (needs bootstrapping the account) and parse last reported number
- Fetch latest vaccincation number from country specific source
- Fetch population from wikimedia
- Tweet status update
