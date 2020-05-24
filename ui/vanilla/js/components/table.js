class Table {
  constructor(columns, rows) {
    this.columns = columns;
    this.rows = rows;
  }

  render(container = null) {
    var html = `
      <table>
        <thead>
          <tr>
            ${this.columns.map((column_title, i) => `<th>${column_title}</th>`).join('')}
          </tr>
        </thead>
        <tbody>
          ${this.rows.map(
              (row_data, i) => 
                `<tr>
                  ${row_data.map((cell_data, j) => 
                    `<td>${cell_data}</td>`).join('')}
                </tr>`
            ).join('')
          }
        </tbody>
      </table>
    `;

    if (container) {
      container.insertAdjacentHTML('afterbegin', html);
    }

    return html;
  }
}