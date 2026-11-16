import { isOverdue, refOf, selectorOf, type Entry } from '../types'

interface Props {
  tasks: Entry[]
  today: string
  selectedRef: string | null
  onSelect: (ref: string) => void
}

export function TaskTable({ tasks, today, selectedRef, onSelect }: Props) {
  if (tasks.length === 0) {
    return <p className="empty">no tasks</p>
  }
  return (
    <table className="tasks">
      <thead>
        <tr>
          <th>project</th>
          <th>id</th>
          <th className="num">P</th>
          <th>status</th>
          <th>due</th>
          <th>title</th>
        </tr>
      </thead>
      <tbody>
        {tasks.map((t) => {
          const ref = refOf(t)
          return (
            <tr
              key={ref}
              className={`${ref === selectedRef ? 'selected' : ''} ${t.status}`}
              onClick={() => onSelect(ref)}
            >
              <td className="project">{selectorOf(t)}</td>
              <td className="id">{t.id}</td>
              <td className={`num p${t.priority}`}>{t.priority}</td>
              <td>
                <span className={`status status-${t.status}`}>{t.status}</span>
              </td>
              <td className={isOverdue(t, today) ? 'overdue' : ''}>{t.due ?? ''}</td>
              <td className="title">
                {t.title}
                {t.tags && t.tags.length > 0 && (
                  <span className="tags">
                    {t.tags.map((tag) => (
                      <span key={tag} className="tag">
                        #{tag}
                      </span>
                    ))}
                  </span>
                )}
              </td>
            </tr>
          )
        })}
      </tbody>
    </table>
  )
}
