// src/components/forms/CaseFormsGrid.tsx
import "./CaseFormsGrid.css";

const forms = [
  {
    title: "VHF Case Investigation Form",
    description: "Report and track VHF cases",
    href: "/forms/vhf",
  },
  {
    title: "M-Pox Case Investigation Form",
    description: "Report and track M-Pox cases",
    href: "/forms/mpox",
  },
  {
    title: "Measles Case Investigation Form",
    description: "Report and track Measles cases",
    href: "/forms/measles",
  },
  {
    title: "Polio Case Investigation Form",
    description: "Report and track Polio cases",
    href: "/forms/polio",
  },
  {
    title: "Other Alerts",
    description: "Report and track other disease cases",
    href: "/forms/other",
  },
];

export default function CaseFormsGrid() {
  return (
    <section className="case-forms">
      <h2 className="section-title">Case Investigation Forms</h2>

      <div className="case-forms__grid">
        {forms.map((form) => (
          <a key={form.title} href={form.href} className="case-form-card">
            <h3>{form.title}</h3>
            <p>{form.description}</p>
          </a>
        ))}
      </div>
    </section>
  );
}
