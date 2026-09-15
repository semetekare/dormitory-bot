import { IconButton } from "@maxhub/max-ui";
import { LuCircleArrowLeft } from 'react-icons/lu';
import { useNavigate } from "react-router";

const BackButton = () => {
    const navigate = useNavigate();

    return (
        <IconButton
            mode="secondary"
            appearance="neutral"
            style={{ zIndex: 1000 }}
            onClick={() => navigate(-1)}
        >
            <LuCircleArrowLeft style={{ fontSize: '2em' }} />
        </IconButton>
    );
};

export default BackButton;