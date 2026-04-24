import React, { useEffect, useState, useCallback, useRef } from 'react';
import PrintPreview from './PrintPreview';

const DynamicForm: React.FC<any> = ({ doctypeName, docName, tenantID, token }) => {
  const [meta, setMeta] = useState<any>(null);
  const [formData, setFormData] = useState<any>({});
  const [workflow, setWorkflow] = useState<any>(null);
  const [isDirty, setIsDirty] = useState(false);
  const [hiddenFields, setHiddenFields] = useState<string[]>([]);
  const [fieldProps, setFieldProps] = useState<any>({});
  const [showPrint, setShowPrint] = useState(false);

  // --- ENGINE: SCRIPT RUNNER ---
  const runScriptEvent = (eventName: string, params: any = {}) => {
    if (!meta?.client_script) return;
    try {
      const frm = {
        doc: formData,
        set_value: (fieldname: string, val: any) => handleChange(fieldname, val),
        toggle_display: (fieldname: string, show: boolean) => {
          setHiddenFields(prev => show ? prev.filter(f => f !== fieldname) : [...new Set([...prev, fieldname])]);
        },
        set_df_property: (fieldname: string, prop: string, val: any) => {
          setFieldProps((prev: any) => ({ ...prev, [fieldname]: { ...(prev[fieldname] || {}), [prop]: val } }));
        },
        refresh: () => refresh(),
        ...params
      };
      const scriptFunc = new Function('frm', meta.client_script);
      const events = scriptFunc(frm) || {};
      if (events[eventName]) events[eventName](frm);
    } catch (e) { console.error("Client Script Error:", e); }
  };

  // --- ENGINE: FETCHING LOGIC ---
  const applyFetchLogic = async (fieldname: string, selectedValue: any, currentMeta: any, isTable: boolean = false, rowIndex: number = -1, tableFieldname: string = "") => {
    if (!selectedValue || !currentMeta) return;
    const fetchFields = currentMeta.fields.filter((f: any) => f.fetch_from?.startsWith(`${fieldname}.`));
    if (fetchFields.length === 0) return;
    const sourceField = currentMeta.fields.find((f: any) => f.name === fieldname);
    if (!sourceField || !sourceField.options) return;
    const res = await fetch(`/api/v1/resource/${sourceField.options}/${selectedValue}`, {
      headers: { 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` }
    });
    if (res.ok) {
      const sourceDoc = await res.json();
      setFormData((prev: any) => {
        let newData = { ...prev };
        if (isTable) {
          const newRows = [...(prev[tableFieldname] || [])];
          fetchFields.forEach((f: any) => {
            const sourceKey = f.fetch_from.split('.')[1];
            newRows[rowIndex] = { ...newRows[rowIndex], [f.name]: sourceDoc[sourceKey] };
          });
          newData[tableFieldname] = newRows;
          return applyCalculations(newData, currentMeta, true, rowIndex, tableFieldname);
        } else {
          const updates: any = {};
          fetchFields.forEach((f: any) => {
            const sourceKey = f.fetch_from.split('.')[1];
            updates[f.name] = sourceDoc[sourceKey];
          });
          newData = { ...newData, ...updates };
          return applyCalculations(newData, meta);
        }
      });
    }
  };

  // --- ENGINE: CALCULATION LOGIC ---
  const applyCalculations = (currentData: any, currentMeta: any, isTable: boolean = false, rowIndex: number = -1, tableFieldname: string = "") => {
    let updatedData = { ...currentData };
    const evaluate = (formula: string, context: any) => {
      try {
        const sanitizedFormula = formula.replace(/[a-zA-Z_][a-zA-Z0-9_]*/g, (match) => {
          return context[match] !== undefined ? context[match] : match;
        });
        return eval(sanitizedFormula);
      } catch (e) { return 0; }
    };
    if (isTable) {
      const rows = [...(updatedData[tableFieldname] || [])];
      if (rowIndex >= 0) {
        const row = rows[rowIndex];
        currentMeta.fields.forEach((f: any) => {
          if (f.formula && !f.formula.includes("(")) row[f.name] = evaluate(f.formula, row);
        });
      }
      if (meta) {
        meta.fields.forEach((hf: any) => {
           if (hf.formula && hf.formula.includes(`${tableFieldname}.`)) {
             if (hf.formula.startsWith("sum(")) {
                const targetCol = hf.formula.match(/\.(.*)\)/)?.[1];
                if (targetCol) updatedData[hf.name] = rows.reduce((acc, r) => acc + (parseFloat(r[targetCol]) || 0), 0);
             }
           }
        });
      }
      updatedData[tableFieldname] = rows;
    } else {
      currentMeta.fields.forEach((f: any) => {
        if (f.formula && !f.formula.includes("(")) updatedData[f.name] = evaluate(f.formula, updatedData);
      });
    }
    return updatedData;
  };

  const handleChange = (field: string, val: any) => {
    setFormData((prev: any) => {
      const newData = { ...prev, [field]: val };
      return applyCalculations(newData, meta);
    });
    setIsDirty(true);
    applyFetchLogic(field, val, meta);
    runScriptEvent(`${field}_on_change`, { fieldname: field, value: val });
  };

  // --- COMPONENTS ---
  const SmartLinkField = ({ field, value, onChange, readOnly }: any) => {
    const [search, setSearch] = useState(value || '');
    const [results, setResults] = useState<any[]>([]);
    const [isOpen, setIsOpen] = useState(false);
    const [loading, setLoading] = useState(false);
    useEffect(() => {
      if (!isOpen || readOnly) return;
      const t = setTimeout(() => {
        setLoading(true);
        fetch(`/api/v1/resource/${field.options}?q=${search}&limit=5`, {
          headers: { 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` }
        }).then(res => res.json()).then(data => {
          setResults(data.data || []);
          setLoading(false);
        });
      }, 300);
      return () => clearTimeout(t);
    }, [search, field.options, isOpen, readOnly]);
    return (
      <div className="relative group">
        <input placeholder={`Search ${field.options}...`} value={isOpen ? search : (value || '')} onChange={(e) => { setSearch(e.target.value); setIsOpen(true); }} onFocus={() => setIsOpen(true)} onBlur={() => setTimeout(() => setIsOpen(false), 200)} className="w-full border p-2 rounded-lg focus:ring-2 focus:ring-indigo-500 outline-none transition text-sm bg-blue-50/10 focus:bg-white border-blue-100 disabled:bg-gray-50" disabled={readOnly} />
        {isOpen && !readOnly && <div className="absolute z-50 w-full mt-1 bg-white border rounded-xl shadow-xl max-h-60 overflow-y-auto border-indigo-100">{loading && <div className="p-3 text-xs text-gray-400 italic animate-pulse">Searching...</div>}{results.map((r: any) => (<div key={r.name} onClick={() => { onChange(r.name); setSearch(r.name); setIsOpen(false); }} className="p-3 hover:bg-indigo-50 cursor-pointer flex flex-col border-b last:border-b-0"><span className="text-sm font-bold text-gray-800">{r.name}</span></div>))}</div>}
      </div>
    );
  };

  const ChildTableGrid = ({ field, value, readOnly }: any) => {
    const [childMeta, setChildMeta] = useState<any>(null);
    const rows = value || [];
    useEffect(() => {
      fetch(`/api/v1/meta/${field.options}`, { headers: { 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` } }).then(res => res.json()).then(data => setChildMeta(data));
    }, [field.options]);
    const updateRow = (idx: number, col: string, val: any) => {
      setFormData((prev: any) => {
        const newRows = [...(prev[field.name] || [])];
        newRows[idx] = { ...newRows[idx], [col]: val };
        const newData = { ...prev, [field.name]: newRows };
        return applyCalculations(newData, childMeta, true, idx, field.name);
      });
      setIsDirty(true);
      applyFetchLogic(col, val, childMeta, true, idx, field.name);
      runScriptEvent(`${field.name}_on_change`, { idx, row: rows[idx], fieldname: col, value: val });
    };
    const addRow = () => {
      setFormData((prev: any) => {
        const newRow: any = {};
        childMeta?.fields.forEach((f: any) => { if(f.default) newRow[f.name] = f.default });
        return { ...prev, [field.name]: [...(prev[field.name] || []), newRow] };
      });
    };
    const removeRow = (idx: number) => {
      setFormData((prev: any) => {
        const newRows = (prev[field.name] || []).filter((_: any, i: number) => i !== idx);
        const newData = { ...prev, [field.name]: newRows };
        return applyCalculations(newData, childMeta, true, -1, field.name);
      });
    };
    if (!childMeta) return <div className="p-4 animate-pulse text-xs text-gray-400">Loading Table...</div>;
    return (
      <div className="mt-4 border rounded-xl overflow-hidden bg-white shadow-inner">
        <table className="w-full text-left">
          <thead className="bg-gray-50 border-b">
            <tr>
              <th className="p-3 text-[10px] font-black text-gray-400 uppercase w-12 text-center">#</th>
              {childMeta.fields.filter((f: any) => f.in_list_view !== false && !['Section Break', 'Column Break'].includes(f.fieldtype)).map((f: any) => (<th key={f.name} className="p-3 text-[10px] font-black text-gray-400 uppercase tracking-widest">{f.label}</th>))}
              {!readOnly && <th className="p-3 w-12"></th>}
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {rows.map((row: any, idx: number) => (
              <tr key={idx} className="hover:bg-indigo-50/20 transition group">
                <td className="p-3 text-center text-[10px] font-bold text-gray-300">{idx + 1}</td>
                {childMeta.fields.filter((f: any) => f.in_list_view !== false && !['Section Break', 'Column Break'].includes(f.fieldtype)).map((f: any) => (
                  <td key={f.name} className="p-2">
                    <FieldRenderer field={f} value={row[f.name]} onChange={(val: any) => updateRow(idx, f.name, val)} readOnly={readOnly || f.read_only} />
                  </td>
                ))}
                {!readOnly && <td className="p-2 text-center"><button onClick={() => removeRow(idx)} className="text-gray-200 hover:text-red-500 transition font-bold text-xl leading-none">×</button></td>}
              </tr>
            ))}
          </tbody>
        </table>
        {!readOnly && <button onClick={addRow} className="w-full p-4 text-[10px] font-black text-indigo-500 uppercase hover:bg-indigo-50 transition border-t tracking-widest">+ Add Row</button>}
      </div>
    );
  };

  const FieldRenderer = ({ field, value, onChange, readOnly }: any) => {
    const props = fieldProps[field.name] || {};
    const isHidden = hiddenFields.includes(field.name) || props.hidden;
    if (isHidden) return null;
    const commonClasses = "w-full border p-2 rounded-lg focus:ring-2 focus:ring-indigo-500 outline-none transition text-sm bg-white disabled:bg-gray-50 disabled:text-gray-400 border-gray-200";
    const finalReadOnly = readOnly || props.read_only;
    if (field.fieldtype === 'Table') return <ChildTableGrid field={field} value={value} readOnly={finalReadOnly} />;
    if (field.fieldtype === 'Link') return <SmartLinkField field={field} value={value} onChange={onChange} readOnly={finalReadOnly} />;
    switch (field.fieldtype) {
      case 'Select':
        return (<select value={value || ''} onChange={(e) => onChange(e.target.value)} className={commonClasses} disabled={finalReadOnly}><option value="">Select...</option>{field.options?.split('\n').map((opt: string) => <option key={opt} value={opt}>{opt}</option>)}</select>);
      case 'Check':
        return <div className="flex items-center h-10 ml-1"><input type="checkbox" checked={!!value} onChange={(e) => onChange(e.target.checked)} className="h-6 w-6 rounded-lg border-gray-300 text-indigo-600 cursor-pointer" disabled={finalReadOnly} /></div>;
      case 'Date':
        return <input type="date" value={value || ''} onChange={(e) => onChange(e.target.value)} className={commonClasses} disabled={finalReadOnly} />;
      case 'Currency':
    case 'Float':
        return <input type="number" step="0.01" value={value || 0} onChange={(e) => onChange(parseFloat(e.target.value))} className={commonClasses + " text-right font-mono font-bold text-gray-700"} disabled={finalReadOnly} />;
      default:
        return <input type="text" value={value || ''} onChange={(e) => onChange(e.target.value)} className={commonClasses} disabled={finalReadOnly} />;
    }
  };

  // --- ACTIONS ---
  const refresh = useCallback(() => {
    const headers = { 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` };
    if (docName) {
      fetch(`/api/v1/resource/${doctypeName}/${docName}`, { headers }).then(res => res.json()).then(data => {
        setFormData(data);
        setIsDirty(false);
        runScriptEvent('refresh');
      });
      fetch(`/api/v1/resource/${doctypeName}/${docName}/workflow`, { headers }).then(res => res.json()).then(data => setWorkflow(data));
    }
  }, [doctypeName, docName, tenantID, token]);

  useEffect(() => {
    const headers = { 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` };
    fetch(`/api/v1/meta/${doctypeName}`, { headers }).then(res => res.json()).then(data => {
      setMeta(data);
      runScriptEvent('setup');
    });
    refresh();
  }, [doctypeName, docName, refresh]);

  const [showCreateMenu, setShowCreateMenu] = useState(false);

  const handleMap = async (targetDT: string) => {
    const res = await fetch(`/api/v1/map/${doctypeName}/${docName}/${targetDT}`, {
      headers: { 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` }
    });
    if (res.ok) {
      const mappedData = await res.json();
      // Logic to switch view to a NEW form with mappedData
      // For now, we update the current form state to simulate a new document
      setFormData(mappedData);
      // setDoctypeName(targetDT); // This would require parent state update
      alert(`Mapped to ${targetDT}. Remember to SAVE.`);
    }
  };

  const onSave = async () => {
    runScriptEvent('before_save');
    const method = docName ? 'PUT' : 'POST';
    const url = docName ? `/api/v1/resource/${doctypeName}/${docName}` : `/api/v1/resource/${doctypeName}`;
    const res = await fetch(url, {
      method,
      headers: { 'Content-Type': 'application/json', 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` },
      body: JSON.stringify(formData)
    });

    const requestID = res.headers.get('X-Request-ID');

    if (res.ok) {
      alert("Saved!");
      refresh();
      runScriptEvent('after_save');
    } else {
      const err = await res.json();
      alert(`Error: ${err.error}\n\nTrace ID: ${requestID}\n(Please provide this ID to your administrator)`);
    }
  };

  if (!meta) return <div className="p-20 text-center animate-pulse text-indigo-600 font-black tracking-widest">INITIALIZING ENGINE...</div>;

  const sections: any[] = [];
  let currentSection: any = { label: 'General Info', columns: [[]] };
  let currentColIndex = 0;
  meta.fields.forEach((field: any) => {
    if (field.fieldtype === 'Section Break') {
      if (currentSection.columns[0].length > 0) sections.push(currentSection);
      currentSection = { label: field.label || 'Details', columns: [[]] };
      currentColIndex = 0;
    } else if (field.fieldtype === 'Column Break') {
      currentSection.columns.push([]);
      currentColIndex++;
    } else {
      currentSection.columns[currentColIndex].push(field);
    }
  });
  sections.push(currentSection);

  return (
    <div className="max-w-7xl mx-auto my-8 bg-gray-50 min-h-screen pb-20">
       <div className="sticky top-0 z-40 bg-white/80 backdrop-blur-md border-b px-8 py-4 flex justify-between items-center shadow-sm">
        <div className="flex items-center space-x-4">
          <div className="p-2 bg-indigo-600 rounded-xl text-white font-black text-xl">GE</div>
          <div>
            <h1 className="text-xl font-black text-gray-900 uppercase tracking-tight">{meta.name} <span className="text-gray-300 font-light"># {docName || 'New'}</span></h1>
            <div className="flex space-x-2 mt-1">
              <span className="text-[10px] font-black px-2 py-0.5 rounded-full uppercase bg-orange-100 text-orange-700">{formData.docstatus === 1 ? 'Submitted' : 'Draft'}</span>
              {workflow?.workflow_active && <span className="text-[10px] font-black px-2 py-0.5 rounded-full uppercase bg-blue-100 text-blue-700">State: {workflow.current_state}</span>}
            </div>
          </div>
        </div>
        <div className="flex space-x-3">
          {docName && meta?.allow_map_to?.length > 0 && (
            <div className="relative">
              <button 
                onClick={() => setShowCreateMenu(!showCreateMenu)}
                className="px-4 py-2 bg-indigo-50 text-indigo-600 rounded-xl font-bold hover:bg-indigo-100 transition border border-indigo-100"
              >
                Create ▼
              </button>
              {showCreateMenu && (
                <div className="absolute right-0 mt-2 w-48 bg-white border rounded-xl shadow-xl z-50 overflow-hidden">
                  {meta.allow_map_to.map((target: string) => (
                    <button 
                      key={target}
                      onClick={() => { handleMap(target); setShowCreateMenu(false); }}
                      className="w-full text-left px-4 py-3 text-sm font-bold text-gray-700 hover:bg-indigo-50 transition border-b last:border-b-0"
                    >
                      New {target}
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}
          {docName && <button onClick={() => setShowPrint(true)} className="px-4 py-2 bg-white border border-gray-200 text-gray-500 rounded-xl hover:text-indigo-600 transition">Print</button>}
          {isDirty && <button onClick={onSave} className="px-6 py-2.5 bg-indigo-600 text-white text-sm font-black rounded-xl shadow-lg active:scale-95 transition">Save</button>}
          {workflow?.transitions?.map((t: any) => <button key={t.action} onClick={async () => {
            const res = await fetch(`/api/v1/resource/${doctypeName}/${docName}/workflow`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json', 'X-GoERP-Tenant': tenantID, 'Authorization': `Bearer ${token}` },
              body: JSON.stringify({ action: t.action })
            });
            if (res.ok) refresh();
          }} className="px-6 py-2.5 bg-white border border-gray-200 text-gray-700 text-sm font-black rounded-xl hover:text-indigo-600 transition shadow-sm"> {t.action} </button>)}
        </div>
      </div>

      <div className="p-8 space-y-10">
        {sections.map((section, sIdx) => (
          <div key={sIdx} className="bg-white rounded-3xl p-10 border border-gray-100 shadow-sm hover:shadow-xl transition-all">
            <h2 className="text-xs font-black text-indigo-600 uppercase tracking-widest mb-8 flex items-center"><span className="w-12 h-[2px] bg-indigo-100 mr-4"></span>{section.label}</h2>
            <div className="flex flex-wrap -mx-6">
              {section.columns.map((col: any[], cIdx: number) => (
                <div key={cIdx} className="flex-1 min-w-[350px] px-6 space-y-8">
                  {col.map(f => {
                    const props = fieldProps[f.name] || {};
                    const isHidden = hiddenFields.includes(f.name) || props.hidden;
                    if (isHidden) return null;
                    return (
                      <div key={f.name} className="flex flex-col">
                         <label className="text-[10px] font-black text-gray-400 uppercase tracking-widest mb-2 ml-1">{props.label || f.label}</label>
                         <FieldRenderer field={f} value={formData[f.name]} onChange={(val: any) => handleChange(f.name, val)} readOnly={f.read_only || formData.docstatus === 1} />
                      </div>
                    );
                  })}
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>

      {showPrint && (
        <div className="fixed inset-0 z-50 overflow-auto bg-black bg-opacity-75 flex flex-col">
          <button onClick={() => setShowPrint(false)} className="absolute top-4 right-10 text-white text-4xl font-light hover:text-indigo-400 transition">×</button>
          <PrintPreview doctype={doctypeName} name={docName} tenantID={tenantID} token={token} />
        </div>
      )}
    </div>
  );
};

export default DynamicForm;
