import {ComboBox, Modal, TextArea} from "@carbon/react";
import {
    type Theme, type ThemeElement,
    useGetThemesQuery,
    useLazyGetThemeElementsQuery
} from "../../datavisualizer/modals/data-model/data-model.ts";
import {useEffect, useState} from "react";
import {useCreateIssueMutation} from "./issue-modal.ts";
import {IssueTypes} from "../constants.ts";
import {useSelector} from "react-redux";
import {selectUser} from "../../../../store/auth/auth.selectors.ts";

export const IssueModal = ({onClose}) => {
    const [datasets, setDatasets] = useState<Theme[] | undefined>([]);
    const [selectedDataset, setSelectedDataset] = useState("");
    const [dataElement, setDataElement] = useState<ThemeElement[]>([]);
    const [description, setDescription] = useState('');
    const [selectedIssueType, setSelectedIssueType] = useState('');
    const [selectedDataElement, setSelectedDataElement] = useState("");
    const { data: themes, isLoading: isLoadingThemes, error } = useGetThemesQuery();
    const [ triggerGetTheme ] = useLazyGetThemeElementsQuery();
    const [createIssue, { isLoading }] = useCreateIssueMutation();
    const user = useSelector(selectUser);

    useEffect(() => {
        if (!isLoadingThemes) {
            setDatasets(themes);
        }
        if (error) {
            console.error("Error Encountered while fetching datasets:: " + error)
        }
    }, [error, isLoadingThemes, themes]);

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

    const handleSubmit = async (formData) => {
        try {
            await createIssue(formData).unwrap();
            onClose();
        } catch (err) {
            console.error("Failed to save the issue: ", err);
        }
    };

    return (
        <Modal
            aria-label="issue-modal"
            modalHeading="Register a New Issue"
            open
            primaryButtonText={isLoading ? "Saving..." : "Submit Issue"}
            secondaryButtonText="Cancel"
            onRequestClose={onClose}
            onRequestSubmit={() => handleSubmit({
                "dataset": selectedDataset,
                "data_element": selectedDataElement,
                "org_unit": "Kampala",
                "issue": description,
                "issue_type": selectedIssueType,
                "reported_by": user?.username
            })}
        >
            <p style={{ marginBottom: '2rem' }}>
                Register a new issue relating to any data anomalies, the causes to it if they are known.
            </p>
            <div style={{ marginBottom: '24px' }}>
                <ComboBox
                    data-modal-primary-focus
                    allowCustomValue
                    autoAlign
                    id="org-unit-combobox"
                    onChange={() => {}}
                    items={[]}
                    titleText="Organization Unit"
                />
            </div>

            <div style={{ marginBottom: '24px' }}>
                <ComboBox
                    allowCustomValue
                    autoAlign
                    id="data-set-combobox"
                    onChange={onChangeSelectedDataSet}
                    items={datasets?.map(item => item?.theme_name) ?? []}
                    titleText="Datasets"
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
                />
            </div>
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