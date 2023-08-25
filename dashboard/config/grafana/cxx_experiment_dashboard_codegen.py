from mako.template import Template
from mako.runtime import Context


def main():
    tmpl_filename = (
        "dashboard/config/grafana/cxx_experiment_dashboard_codegen.json.template"
    )
    out_filename = "dashboard/config/grafana/dashboards/psm/cxx_experiments.json"
    try:
        cxx_template = Template(filename=tmpl_filename, format_exceptions=True)
    except NameError:
        print("Mako template is not installed. Skipping HTML report generation.")
        return
    except IOError as e:
        print(("Failed to find the template %s: %s" % (tmpl_filename, e)))
        return

    experiments = [
        {
            "name": "baseline",
            "8core_table": "ci_master_results_8core",
            "32core_table": "ci_master_results_32core",
        },
        {
            "name": "event_engine_listener",
            "8core_table": "results_8core_event_engine_listener",
            "32core_table": "results_32core_event_engine_listener",
        },
        {
            "name": "event_engine_client",
            "8core_table": "results_8core_event_engine_client",
            "32core_table": "results_32core_event_engine_client",
        },
        {
            "name": "event_engine_client,event_engine_listener",
            "8core_table": "results_8core_event_engine_client__event_engine_listener",
            "32core_table": "results_32core_event_engine_client__event_engine_listener",
        },
    ]
    series_overrides = """[
          {
            "alias": "baseline",
            "color": "#8AB8FF"
          },
          {
            "alias": "event_engine_listener",
            "color": "#FFEE52"
          },
          {
            "alias": "event_engine_client",
            "color": "#96D98D"
          },
          {
            "alias": "event_engine_client,event_engine_listener",
            "color": "#F2495C"
          }
        ]"""

    args = {
        "experiments": experiments,
        "series_overrides": series_overrides,
    }
    with open(out_filename, "w") as outfile:
        cxx_template.render_context(Context(outfile, **args))


if __name__ == "__main__":
    main()
