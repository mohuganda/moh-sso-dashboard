import { Modal } from "@carbon/react";

export default function ClearDataConfirmationModal({
    isClearModalOpen,
    setIsClearModalOpen,
    handleConfirmClear,
}){

    return (
        <>
            <Modal
                open={isClearModalOpen}
                modalHeading="Clear all selections?"
                primaryButtonText="Clear all"
                secondaryButtonText="Cancel"
                danger
                onRequestClose={() => setIsClearModalOpen(false)}
                onRequestSubmit={handleConfirmClear}
            >
                <p style={{ padding: '0 1rem' }}>
                    This will reset all selected data elements, periods, and organisation units. This action cannot be undone.
                </p>
            </Modal>
        </>
    );
}
