import {ComboBox, Modal, TextArea, Search, TreeView} from "@carbon/react";
import {
    type Theme, type ThemeElement,
    useGetThemesQuery,
    useLazyGetThemeElementsQuery
} from "../../datavisualizer/modals/data-model/data-model.ts";
import {useCallback, useEffect, useState} from "react";
import {useCreateIssueMutation, useUpdateIssueMutation} from "./issue-modal.ts";
import {IssueTypes, Priority} from "../constants.ts";
import {useSelector} from "react-redux";
import {selectUser} from "../../../../store/auth/auth.selectors.ts";
import type {Issue} from "../issue-tracker.tsx";
import { ChevronDown, ChevronUp } from "@carbon/react/icons";
import {useGetHierarchyQuery} from "../../datavisualizer/modals/orgunit/org-unit.ts";
import {OrgUnitNode} from "../function.tsx";

export const IssueModal = ({ onClose, selectedIssue }: { onClose: () => void, selectedIssue?: Issue | null }) => {
    const isEdit = !!selectedIssue;
    const [datasets, setDatasets] = useState<Theme[] | undefined>([]);
    const [selectedDataset, setSelectedDataset] = useState(selectedIssue?.dataset ?? "");
    const [dataElement, setDataElement] = useState<ThemeElement[]>([]);
    const [description, setDescription] = useState(selectedIssue?.issue ?? '');
    const [selectedIssueType, setSelectedIssueType] = useState(selectedIssue?.issue_type ?? '');
    const [selectedDataElement, setSelectedDataElement] = useState(selectedIssue?.data_element ?? "");
    const [priority, setPriority] = useState(selectedIssue?.priority ?? "");
    const [severity, setSeverity] = useState(selectedIssue?.severity ?? "");
    const { data: themes, isLoading: isLoadingThemes, error } = useGetThemesQuery();
    const [ triggerGetTheme ] = useLazyGetThemeElementsQuery();
    const [updateIssue, { isLoading: isUpdating }] = useUpdateIssueMutation();
    const [createIssue, { isLoading: isCreating }] = useCreateIssueMutation();
    const user = useSelector(selectUser);
    const [isOrgExpanded, setIsOrgExpanded] = useState(false);
    const [selectedOrgUnit, setSelectedOrgUnit] = useState(selectedIssue?.org_unit ?? "");
    const [orgUnits, setOrgUnits] = useState<any>({});
    const { data: hierarchyData, isLoading, error:hierarchyDataError} = useGetHierarchyQuery();
    const [orgSearchTerm, setOrgSearchTerm] = useState("");

    useEffect(() => {
        if (!isLoadingThemes) {
            setDatasets(themes);
        }
        if (error) {
            console.error("Error Encountered while fetching datasets:: " + error)
        }
    }, [error, isLoadingThemes, themes]);

    useEffect(() => {
        if(!isLoading) {
            setOrgUnits(hierarchyData);
        }
        if (hierarchyDataError) {
            console.error("Error Encountered while fetching org units:: " + hierarchyDataError)
        }
    },[error, hierarchyData, hierarchyDataError, isLoading]);

    const onChangeSelectedDataSet = async (event) => {
        const theme = event?.selectedItem;
        setSelectedDataset(theme);
        if (!theme) return;

        const theme_id = themes?.find(item => item?.theme_name === theme)?.theme_id;
        if (theme_id) {
            try {
                const data = await triggerGetTheme(theme_id).unwrap();
                setDataElement(data);
            } catch (error) {
                console.error("Error Encountered while fetching data elements:: " + error);
            }
        } else {
            console.warn("No theme_id found for the selected theme name.");
        }
    };

    const onChangeSelectedDataElement = (event) => {
        setSelectedDataElement(event?.selectedItem);
    }

    const onChangeIssueType = (event) => {
        setSelectedIssueType(event?.selectedItem);
    }

    const handleTextChange = (event) => {
        setDescription(event.target.value);
    };

    const onChangePriority = (event) => {
        setPriority(event?.selectedItem);
    }

    const onChangeSeverity = (event) => {
        setSeverity(event?.selectedItem);
    }

    const handleSubmit = async (formData) => {
        try {
            if (isEdit) {
                await updateIssue({ id: selectedIssue?.issue_code, body: formData }).unwrap();
            } else {
                await createIssue(formData).unwrap();
            }
            onClose();
        } catch (err) {
            console.error("Failed to save the issue: ", err);
        }
    };

    const nodeMatchesSearch = useCallback((node: any, term: string): boolean => {
        if (node.name.toLowerCase().includes(term.toLowerCase())) {
            return true;
        }
        if (node.children) {
            return node.children.some((child: any) => nodeMatchesSearch(child, term));
        }
        return false;
    },[]);

    const handleSelect = (name: string) => {
        setSelectedOrgUnit(name);
        setOrgSearchTerm("");
        setTimeout(() => setIsOrgExpanded(false), 150);
    };

    const renderRecursive = (nodes: any[]) => {
        if (!nodes || !Array.isArray(nodes)) return [];

        return nodes
            .filter(node => !orgSearchTerm || nodeMatchesSearch(node, orgSearchTerm))
            .map(node => (
                <OrgUnitNode
                    key={node.id}
                    node={node}
                    searchTerm={orgSearchTerm}
                    selectedOrgUnit={selectedOrgUnit}
                    onSelect={handleSelect}
                    renderRecursive={renderRecursive}
                />
            ));
    };

    return (
        <Modal
            aria-label="issue-modal"
            open
            modalHeading={isEdit ? `Edit Issue: ${selectedIssue.issue_code}` : "Register a New Issue"}
            primaryButtonText={(isCreating || isUpdating) ? "Saving..." : "Submit Issue"}
            secondaryButtonText="Cancel"
            onRequestClose={onClose}
            onRequestSubmit={() => handleSubmit({
                dataset: selectedDataset,
                data_element: selectedDataElement,
                org_unit: selectedOrgUnit,
                issue: description,
                issue_type: selectedIssueType,
                reported_by: user?.username,
                updated_by: isEdit ? user?.username : "",
                priority: isEdit ? priority : "",
                severity: isEdit ? severity : "",

            })}
        >
            <p style={{ marginBottom: '2rem' }}>
                Register a new issue relating to any data anomalies, the causes to it if they are known.
            </p>
            <div style={{ marginBottom: '24px' }}>
                <p className="cds--label">Organisation Unit</p>
                <div
                    onClick={() => setIsOrgExpanded(!isOrgExpanded)}
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        padding: '0 1rem',
                        height: '40px',
                        background: '#f4f4f4',
                        borderBottom: '1px solid #8d8d8d',
                        cursor: 'pointer'
                    }}
                >
                <span style={{ color: selectedOrgUnit ? '#161616' : '#6f6f6f' }}>
                    {selectedOrgUnit || "Select Organisation Unit"}
                </span>
                    {isOrgExpanded ? <ChevronUp /> : <ChevronDown />}
                </div>

                {isOrgExpanded && (
                    <div style={{
                        border: '1px solid #e0e0e0',
                        background: 'white',
                        maxHeight: '250px',
                        overflowY: 'auto',
                        marginTop: '2px',
                        padding: '8px'
                    }}>
                        <Search
                            labelText=""
                            size="md"
                            id="org-search-stable"
                            placeholder="Search for org unit..."
                            value={orgSearchTerm}
                            onChange={(e) => setOrgSearchTerm(e.target.value)}
                            style={{ marginBottom: '8px' }}
                        />
                        <TreeView label="Org Units" hideLabel>
                            {renderRecursive(orgUnits)}
                        </TreeView>
                    </div>
                )}
            </div>
            <div style={{ marginBottom: '24px' }}>
                <ComboBox
                    allowCustomValue
                    autoAlign
                    id="data-set-combobox"
                    onChange={onChangeSelectedDataSet}
                    items={datasets?.map(item => item?.theme_name) ?? []}
                    titleText="Datasets"
                    selectedItem={selectedDataset}
                />
            </div>
            <div style={{ marginBottom: '24px' }}>
                <ComboBox
                    allowCustomValue
                    autoAlign
                    id="data-element-combobox"
                    onChange={onChangeSelectedDataElement}
                    items={dataElement?.map(item => item?.data_element_short_name)}
                    titleText="Data Elements"
                    selectedItem={selectedDataElement}
                />
            </div>
            <div style={{ marginBottom: '24px' }}>
                <ComboBox
                    allowCustomValue
                    autoAlign
                    id="data-element-combobox"
                    onChange={onChangeIssueType}
                    items={IssueTypes}
                    titleText="Issue Type"
                    selectedItem={selectedIssueType}
                />
            </div>
            {isEdit ? (
                <>
                <div style={{ marginBottom: '24px' }}>
                    <ComboBox
                        allowCustomValue
                        autoAlign
                        id="priority-combobox"
                        onChange={onChangePriority}
                        items={Priority}
                        titleText="Priority"
                        selectedItem={priority}
                    />
                </div>
                <div style={{ marginBottom: '24px' }}>
                    <ComboBox
                        allowCustomValue
                        autoAlign
                        id="severity-combobox"
                        onChange={onChangeSeverity}
                        items={Priority}
                        titleText="Severity"
                        selectedItem={severity}
                    />
                </div>
                </>
            ) : null }
            <TextArea
                id="issue-text-area"
                labelText="Issue Description"
                style={{ marginBottom: '24px' }}
                value={description}
                onChange={handleTextChange}
                rows={7}
            />
        </Modal>
    )
}