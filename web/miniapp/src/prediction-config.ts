/** Survey inputs — keys must match backend API / Excel schema. */
export const PREDICTION_INPUT_FIELDS = [
  { key: "age", label: "Ваш возраст (полных лет)?" },
  { key: "pregnancy_weeks", label: "Срок беременности (недель)?" },
  {
    key: "kpu_index",
    label:
      "Индекс КПУ (количество кариозных, пломбированных и удаленных зубов) — по данным вашего стоматолога:",
  },
  { key: "brushing_per_day", label: "Сколько раз в день вы чистите зубы?" },
  {
    key: "dentist_visit_during_pregnancy",
    label: "Посещали ли вы стоматолога во время беременности?",
  },
  { key: "parent_caries", label: "Был ли кариес у ваших родителей?" },
  { key: "saliva_ph", label: "рН вашей слюны (по результатам анализа)" },
] as const;

/** Predicted outputs returned by the API. */
export const PREDICTION_OUTPUT_FIELDS = [
  {
    key: "child_caries_probability",
    label: "Вероятность развития кариеса у ребенка",
  },
  { key: "risk_group", label: "Группа риска" },
  { key: "action", label: "Действие" },
  { key: "recommendations", label: "Назначение (Рекомендации)" },
] as const;

export type PredictionInputKey = (typeof PREDICTION_INPUT_FIELDS)[number]["key"];
export type PredictionOutputKey =
  (typeof PREDICTION_OUTPUT_FIELDS)[number]["key"];
