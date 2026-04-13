import React from "react";
import "./panel.scss";

interface PanelProps {
    heading: string;
    children: React.ReactNode;
}

const Panel: React.FC<PanelProps> = ({ heading, children }) => {
    return (
        <div className={`panel`}>
            <div className={`heading`}>
                <span>{heading}</span>
            </div>
            {children}
        </div>
    );
};

export default Panel;
