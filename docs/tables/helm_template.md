# Table: helm_template

A template is a file that defines a Kubernetes manifest in a way that is generic enough to allow customization at the time of installation. It can reference variables and functions that are provided by Helm or defined in the chart.

During the installation process, Helm takes the template files in the chart and renders them using the values provided by the user or the defaults defined in the chart's values.yaml file.

## Examples

### Basic info

```sql
select
  name,
  chart_name,
  rendered,
  raw
from
  helm_template
where
  chart_name = 'steampipe-cloud-foundation';
```

### Query templates with values to override the default value

```sql
select
  name,
  chart_name,
  rendered,
  raw
from
  helm_template
where
  chart_name = 'steampipe'
  and vals = '{
    "nfs": {
      "path": "/test/path"
    }
  }';
```

### Query templates with value file to override the default value

```sql
select
  name,
  chart_name,
  rendered,
  raw
from
  helm_template
where
  chart_name = 'steampipe'
  and val_file = '/path/to/value.yaml';
```
