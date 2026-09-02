import type { Statistics as StatisticsData } from "../types";

type StatisticsProps = {
  statistics: StatisticsData;
  formatNumber: (value: number) => string;
};

export function Statistics({ statistics, formatNumber }: StatisticsProps) {
  const values = [
    ["Máximo", formatNumber(statistics.max)],
    ["Mínimo", formatNumber(statistics.min)],
    ["Suma", formatNumber(statistics.sum)],
    ["Promedio", formatNumber(statistics.average)],
    ["¿Alguna matriz es diagonal?", statistics.hasDiagonalMatrix ? "Sí" : "No"]
  ];

  return (
    <section aria-labelledby="statistics-title" className="statistics-panel">
      <h3 id="statistics-title">Estadísticas</h3>
      <dl className="statistics-grid">
        {values.map(([label, value]) => (
          <div className="statistic" key={label}>
            <dt>{label}</dt>
            <dd>{value}</dd>
          </div>
        ))}
      </dl>
    </section>
  );
}
