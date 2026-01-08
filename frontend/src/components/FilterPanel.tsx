import React from "react";

import type {FilterSpec} from "../types";

export type FilterState = Record<
    string,
    {
        operator?: string;
        values: string[];
    }
>;

type Props = {
    filters: FilterSpec[];
    state: FilterState;
    onChange: (next: FilterState) => void;
    onApply: () => void;
    onReset: () => void;
};

export const FilterPanel: React.FC<Props> = ({filters, state, onChange, onApply, onReset}) => {
    return (
        <div className="filters">
            <table className="filter-table">
                <tbody>
                    {filters.map((filter) => {
                        const current = state[filter.id] ?? {values: []};
                        const operators = resolveOperators(filter);
                        const operator = current.operator ?? operators[0] ?? "eq";

                        return (
                            <tr key={filter.id}>
                                <th scope="row" className="filter-label">
                                    {filter.title}
                                </th>
                                <td className="filter-controls">
                                    {operators.length > 1 && (
                                        <select
                                            className="filter-operator"
                                            value={operator}
                                            onChange={(event) =>
                                                updateFilter(state, filter.id, onChange, {
                                                    operator: event.target.value,
                                                })
                                            }
                                        >
                                {operators.map((op) => (
                                    <option key={op} value={op}>
                                        {op || "is"}
                                    </option>
                                ))}
                            </select>
                        )}
                                    {renderFilterInput(filter, current, operator, state, onChange)}
                                </td>
                            </tr>
                        );
                    })}
                    <tr className="filter-row-actions">
                        <th scope="row" className="filter-label">
                            Actions
                        </th>
                        <td className="filter-controls">
                            <div className="filter-actions">
                                <button type="button" onClick={onApply}>
                                    Apply
                                </button>
                                <button type="button" onClick={onReset}>
                                    Reset
                                </button>
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    );
};

function renderFilterInput(
    filter: FilterSpec,
    current: FilterState[string],
    operator: string,
    state: FilterState,
    onChange: (next: FilterState) => void,
) {
    const commonProps = {
        className: "filter-input",
    };
    if (filter.type === "select_one") {
        return (
            <select
                {...commonProps}
                value={current.values[0] ?? ""}
                onChange={(event) =>
                    updateFilter(state, filter.id, onChange, {
                        operator,
                        values: event.target.value ? [event.target.value] : [],
                    })
                }
            >
                <option value="">Select...</option>
                {filter.values?.map((option) => (
                    <option key={option.value} value={option.value}>
                        {option.label}
                    </option>
                ))}
            </select>
        );
    }

    if (filter.type === "select_multi") {
        return (
            <select
                {...commonProps}
                multiple
                value={current.values}
                onChange={(event) =>
                    updateFilter(state, filter.id, onChange, {
                        operator: operator || "in",
                        values: Array.from(event.target.selectedOptions).map((opt) => opt.value),
                    })
                }
            >
                {filter.values?.map((option) => (
                    <option key={option.value} value={option.value}>
                        {option.label}
                    </option>
                ))}
            </select>
        );
    }

    const inputType = resolveInputType(filter.type);
    if (operator === "between") {
        return (
            <>
                <input
                    {...commonProps}
                    type={inputType}
                    placeholder="From"
                    value={current.values[0] ?? ""}
                    onChange={(event) =>
                        updateFilter(state, filter.id, onChange, {
                            operator,
                            values: [event.target.value, current.values[1] ?? ""].filter(
                                (value) => value !== "",
                            ),
                        })
                    }
                />
                <input
                    {...commonProps}
                    type={inputType}
                    placeholder="To"
                    value={current.values[1] ?? ""}
                    onChange={(event) =>
                        updateFilter(state, filter.id, onChange, {
                            operator,
                            values: [current.values[0] ?? "", event.target.value].filter(
                                (value) => value !== "",
                            ),
                        })
                    }
                />
            </>
        );
    }
    return (
        <input
            {...commonProps}
            type={inputType}
            value={current.values[0] ?? ""}
            onChange={(event) =>
                updateFilter(state, filter.id, onChange, {
                    operator,
                    values: event.target.value ? [event.target.value] : [],
                })
            }
        />
    );
}

function updateFilter(
    state: FilterState,
    id: string,
    onChange: (next: FilterState) => void,
    patch: Partial<FilterState[string]>,
) {
    const current = state[id] ?? {values: []};
    onChange({
        ...state,
        [id]: {
            ...current,
            ...patch,
        },
    });
}

export function resolveOperators(filter: FilterSpec): string[] {
    if (filter.operators && filter.operators.length > 0) {
        const hasEq = filter.operators.includes("eq");
        const ops = filter.operators.filter((op) => op !== "eq");
        return hasEq ? [""].concat(ops) : ops;
    }
    if (filter.type === "select_multi") {
        return ["in"];
    }
    return [""];
}

export function emptyFilterState(filters: FilterSpec[]): FilterState {
    return filters.reduce<FilterState>((acc, filter) => {
        acc[filter.id] = {operator: resolveOperators(filter)[0], values: []};
        return acc;
    }, {});
}

export function parseFilterParams(filters: FilterSpec[], search: string): FilterState {
    const filterIndex = filters.reduce<Record<string, FilterSpec>>((acc, filter) => {
        acc[filter.id] = filter;
        return acc;
    }, {});
    const params = new URLSearchParams(search);
    const state: FilterState = emptyFilterState(filters);
    let hasFilters = false;
    for (const [key, value] of params.entries()) {
        if (key === "limit" || key === "offset") continue;
        const dotIndex = key.lastIndexOf(".");
        const id = dotIndex === -1 ? key : key.slice(0, dotIndex);
        const operator = dotIndex === -1 ? undefined : key.slice(dotIndex + 1);
        if (!filterIndex[id]) continue;
        const spec = filterIndex[id];
        hasFilters = true;
        const current = state[id] ?? {operator: resolveOperators(spec)[0], values: []};
        const normalizedValue = normalizeFilterValue(spec, value);
        if (normalizedValue === "") {
            continue;
        }
        state[id] = {
            operator: operator ?? current.operator,
            values: [...current.values, normalizedValue].filter((item) => item !== ""),
        };
    }
    if (!hasFilters) {
        return emptyFilterState(filters);
    }
    return state;
}

export function appendFilters(
    params: URLSearchParams,
    filters: FilterState,
    specs: FilterSpec[] = [],
) {
    const filterIndex = specs.reduce<Record<string, FilterSpec>>((acc, filter) => {
        acc[filter.id] = filter;
        return acc;
    }, {});
    Object.entries(filters).forEach(([id, filter]) => {
        if (!filter.values || filter.values.length === 0) return;
        const spec = filterIndex[id];
        if (!spec) return;
        const operator = filter.operator ?? resolveOperators(spec)[0];
        const omitOperator = operator === "in" && spec.type === "select_multi";
        const key = operator && !omitOperator ? `${id}.${operator}` : id;
        if (operator === "between" || spec.type === "select_multi") {
            filter.values.forEach((value) => {
                if (value !== "") {
                    const serializedValue = serializeFilterValue(spec, value);
                    if (serializedValue !== "") {
                        params.append(key, serializedValue);
                    }
                }
            });
            return;
        }
        if (filter.values[0] !== "") {
            const serializedValue = serializeFilterValue(spec, filter.values[0]);
            if (serializedValue !== "") {
                params.append(key, serializedValue);
            }
        }
    });
}

function resolveInputType(type: string) {
    if (type === "number") return "number";
    if (type === "date") return "date";
    if (type === "datetime") return "datetime-local";
    return "text";
}

function normalizeFilterValue(spec: FilterSpec, value: string): string {
    if (spec.type !== "datetime") {
        return value;
    }
    return fromUnixTimestamp(value);
}

function serializeFilterValue(spec: FilterSpec, value: string): string {
    if (spec.type !== "datetime") {
        return value;
    }
    if (/^\d+$/.test(value)) {
        return value;
    }
    return toUnixTimestamp(value);
}

function fromUnixTimestamp(value: string): string {
    if (!/^\d+$/.test(value)) {
        return value;
    }
    const numberValue = Number(value);
    const tzoffset = (new Date()).getTimezoneOffset() * 60000; //offset in milliseconds
    const localISOTime = (new Date(numberValue * 1000- tzoffset)).toISOString().slice(0, 16);

    console.log(localISOTime); // "2023-09-01T11:00"
    return localISOTime;
}

function toUnixTimestamp(value: string): string {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
        return "";
    }
    return Math.floor(date.getTime() / 1000).toString();
}
