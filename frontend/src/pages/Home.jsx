export default function Home() {
  return (
    <section className="section is-fullheight has-background-light">
      <div
        className="container is-flex is-justify-content-center is-align-items-center"
        style={{ minHeight: '80vh', maxWidth: '600px' }}
      >
        <div className="box has-text-centered p-6">
          <h1 className="title is-2">CCS Management</h1>
          <hr />
          <p className="subtitle is-5 mb-5">
            Please log in to access your dashboard and manage your servers.
          </p>
          <a href="/dashboard" className="button is-medium">
            Go to Dashboard
          </a>
        </div>
      </div>
    </section>
  );
}
