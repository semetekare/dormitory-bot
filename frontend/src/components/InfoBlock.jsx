import '../styles/InfoBlock.css'

const InfoBlock = ({ title, content }) => (
    <div className="info_block mt-1">
        <div className='info_block-title'>
            <span>
                {title}
            </span>
        </div>
        <div className='info_block-content'>
            <span>
                {content}
            </span>
        </div>
    </div>
);

export default InfoBlock;