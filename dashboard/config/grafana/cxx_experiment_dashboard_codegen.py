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
            "name": "work_stealing",
            "8core_table": "results_8core_work_stealing",
            "32core_table": "results_32core_work_stealing",
        },
        {
            "name": "event_engine_listener,work_stealing",
            "8core_table": "results_8core_event_engine_listener__work_stealing",
            "32core_table": "results_32core_event_engine_listener__work_stealing",
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
            "alias": "work_stealing",
            "color": "#E02F44"
          },
          {
            "alias": "event_engine_listener,work_stealing",
            "color": "#96D98D"
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
