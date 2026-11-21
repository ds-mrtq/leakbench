def report_total(rows):
    return sum(row['total'] for row in rows)


def create_app():
    from flask import Flask
    app = Flask(__name__)

    @app.get('/report')
    def report():
        return {'total': report_total([{'total': 3}, {'total': None}])}

    return app


if __name__ == '__main__':
    print(report_total([{'total': 3}, {'total': None}]))
