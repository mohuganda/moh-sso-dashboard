import DataList from "../datavisualizer/components/data-table/data-table.component.tsx";
import {Button} from "@carbon/react";
import {Add} from "@carbon/react/icons";
import "./issue-tracker.scss";
import {useEffect, useState} from "react";
import {IssueModal} from "./Modals/issue-modal.tsx";
import { useGetIssuesQuery} from "./Modals/issue-modal.ts";
import { headers } from "./constants.ts";
import IssueDetail from "./issuedetail/issue-detail.tsx";

export type Issue = {
    issue_id: number;
    issue_code: string;
    dataset: string;
    data_element: string;
    org_unit: string;
    issue: string;
    date_reported: string;
    reported_by: string;
    status: string;
    issue_type: string;
    updated_by: string;
    updated_date: string;
    priority: string;
    severity: string;
    time_period: string;
    time_Period: string;
}

const IssueTracker = () => {
    const [showModal, setShowModal] = useState(false);
    const [issues, setIssues] = useState<Issue[]>([]);
    const { data, isLoading, error } = useGetIssuesQuery();
    const [selectedIssue, setSelectedIssue] = useState<Issue>();
    const [isViewIssueDetail, setIsViewIssueDetail] = useState(false);

    const close = () => {
        setShowModal(false);
    };


    useEffect(() => {
        if (!isLoading) {
            const rows = data?.data?.map(item => ({
                id: item.issue_id.toString(),
                ...item
            }));
            setIssues(rows);
        }
        if (error) {
            console.error("Error Encountered while fetching issues:: " + error)
        }
    }, [data, error, isLoading]);

    const handleIssueClick = (issue) => {
        const selectedItem = issues.find(item => item?.issue_id.toString() === issue?.id);
        if (selectedItem) {
            setSelectedIssue(selectedItem);
            setIsViewIssueDetail(true);
        }
    };

  return (
    <>
        { isViewIssueDetail && selectedIssue ? (
            <IssueDetail selectedIssue={selectedIssue} goToBack={() => setIsViewIssueDetail(false)}/>
        ) : (
            <>
                <div className="dv-toolbar issue-label-container">
                    <div>
                        <span className="issue-label"> Registered Issues </span>
                    </div>
                    <Button
                        size="md"
                        kind="primary"
                        renderIcon={Add}
                        className={`dwh-btn-width`}
                        onClick={()=> setShowModal(true)}
                    >
                        New Issue
                    </Button>
                </div>
                <div className="issue-container">
                    <DataList columns={headers} data={issues} handleIssueClick={handleIssueClick}/>
                </div>
                {
                    showModal && <IssueModal onClose={close}/>
                }
            </>
        )}
    </>
  );
};

export default IssueTracker;
